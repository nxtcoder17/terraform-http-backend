package types

// // EncryptedState encrypted state
//
//	type EncryptedState struct {
//		EncryptedData string `json:"encrypted_data"`
//	}
//
// // StateDocument a state with reference
//
//	type StateDocument struct {
//		Ref       string                 `json:"ref"`
//		Encrypted bool                   `json:"encrypted"`
//		State     map[string]interface{} `json:"state"`
//		Metadata  map[string]interface{} `json:"metadata"`
//	}
//
// // LockDocument a lock with reference
//
//	type LockDocument struct {
//		Ref  string `json:"ref"`
//		Lock Lock   `json:"lock"`
//	}
//
// Lock derived from logging LOCK request body
type Lock struct {
	Created   string
	Path      string
	ID        string
	Operation string
	Info      string
	Who       string
	Version   string
}

type State struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}
