//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

//// RootKey returns the current root key in a thread-safe manner.
//// The returned key is protected by a read lock to ensure concurrent
//// access safety.
//func RootKey() string {
//	rootKeyMu.RLock()
//	defer rootKeyMu.RUnlock()
//
//	return rootKey
//}
//
//// SetRootKey sets the root key that is fetched from SPIKE Keeper.
//func SetRootKey(key string) {
//	rootKeyMu.Lock()
//	defer rootKeyMu.Unlock()
//
//	rootKey = key
//}
