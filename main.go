package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nxtcoder17/ivy"
	"github.com/nxtcoder17/ivy/middleware"
	logger_mw "github.com/nxtcoder17/ivy/middleware/logger"

	"github.com/nxtcoder17/terraform-backend-http/pkg/encryption"
	"github.com/nxtcoder17/terraform-backend-http/store"
)

func ValueAs[T any](v any) (T, bool) {
	x, ok := v.(T)
	if ok {
		return x, true
	}
	return x, false
}

var (
	addr string
	args []string
)

func init() {
	flag.StringVar(&addr, "addr", ":3000", "--addr [host]:<port>")
	flag.Parse()
	args = flag.Args()
}

func serve() {
	logger := slog.Default()

	router := ivy.NewRouter(ivy.WithErrorHandler(func(err error, w http.ResponseWriter, r *http.Request) {
		logger.Error("received", "err", err)
		http.Error(w, err.Error(), 500)
	}))

	router.Use(logger_mw.New(logger_mw.WithLogger(logger)))

	router.Get("/ping", func(c *ivy.Context) error {
		return c.SendString("OK")
	})

	fsStore, err := store.NewFileSystemStore([]byte(os.Getenv("ENCRYPTION_KEY")))
	if err != nil {
		panic(err)
	}

	lockMap := make(map[string]string)

	mw := func(c *ivy.Context) error {
		dir := c.QueryParam("dir")
		if dir == "" {
			return fmt.Errorf("invalid query-param (dir = %q)", dir)
		}

		dir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		c.KV.Set("lock-file", filepath.Join(dir, "terraform-http.lock"))
		c.KV.Set("state-file", filepath.Join(dir, "terraform-http.state.json"))
		return c.Next()
	}

	router.Method("LOCK", "/", mw, func(c *ivy.Context) error {
		b, err := io.ReadAll(c.Body())
		if err != nil {
			return err
		}
		defer c.Body().Close()

		lockfile, _ := ValueAs[string](c.KV.Get("lock-file"))

		logger.Debug("lock:info", "body", string(b))

		lock, err := fsStore.LockState(c, store.LockStateArgs{
			Lockfile: lockfile,
			Body:     b,
		})
		if err != nil {
			return err
		}

		lockMap[lockfile] = lock.ID

		return c.SendStatus(http.StatusOK)
	})

	router.Method("UNLOCK", "/", mw, func(c *ivy.Context) error {
		lockfile, _ := ValueAs[string](c.KV.Get("lock-file"))

		var lock store.Lock
		if err := c.ParseBodyInto(&lock); err != nil {
			return err
		}

		logger.Debug("lock:info", "lock.ID", lock.ID, "lockmap", lockMap)

		if v, ok := lockMap[lockfile]; !ok || v != lock.ID {
			return fmt.Errorf("resource is not locked by this server")
		}

		delete(lockMap, lockfile)

		if err := fsStore.UnlockState(c, store.UnlockStateArgs{
			Lockfile: lockfile,
		}); err != nil {
			return err
		}

		return c.SendStatus(http.StatusOK)
	})

	router.Get("/", mw, func(c *ivy.Context) error {
		statefile, _ := ValueAs[string](c.KV.Get("state-file"))

		b, err := fsStore.ReadState(statefile)
		if err != nil {
			return err
		}
		return c.SendString(string(b))
	})

	router.Post("/", middleware.MustHaveQueryParams("ID"), mw, func(c *ivy.Context) error {
		id := c.QueryParam("ID")
		statefile, _ := ValueAs[string](c.KV.Get("state-file"))

		b, err := io.ReadAll(c.Body())
		if err != nil {
			return err
		}
		defer c.Body().Close()

		logger.Debug("got", "body", string(b), "id", id)

		if err := fsStore.WriteState(statefile, b); err != nil {
			return err
		}

		return c.SendStatus(http.StatusOK)
	})

	router.Delete("/", middleware.MustHaveQueryParams("ID"), mw, func(c *ivy.Context) error {
		// id := c.QueryParam("ID")
		statefile, _ := ValueAs[string](c.KV.Get("state-file"))

		if err := fsStore.DeleteState(statefile); err != nil {
			return err
		}

		return c.SendStatus(http.StatusOK)
	})

	slog.Info("HTTP server starting on", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		panic(err)
	}
}

func main() {
	if len(args) == 0 {
		panic(fmt.Errorf("must have an argument, one of [serve | encrypt | decrypt]"))
	}

	switch args[0] {
	case "serve":
		serve()
	case "encrypt":
		{
			if len(args) > 1 {
			}

			c, err := encryption.NewAESCipher([]byte(os.Getenv("ENCRYPTION_KEY")))
			if err != nil {
				panic(err)
			}

			v, err := c.Encrypt([]byte(args[1]))
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s\n", v)
		}
	case "decrypt":
		{
			c, err := encryption.NewAESCipher([]byte(os.Getenv("ENCRYPTION_KEY")))
			if err != nil {
				panic(err)
			}

			out, err := c.Decrypt([]byte(args[1]))
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s\n", out)
		}
	}
}
