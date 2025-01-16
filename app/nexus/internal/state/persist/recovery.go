//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

//func AsyncPersistRecoveryInfo(meta store.KeyRecoveryData) {
//	be := Backend()
//
//	// TODO: if be == nil, then retry later.
//
//	go func() {
//		ctx, cancel := context.WithTimeout(
//			context.Background(),
//			env.DatabaseOperationTimeout(),
//		)
//		defer cancel()
//
//		fmt.Println("<<<<< BEFORE STORING RECOVERY INFO >>>>>")
//
//		if err := be.StoreKeyRecoveryInfo(ctx, meta); err != nil {
//			log.Log().Warn("asyncPersistRecoveryInfo",
//				"msg", "Failed to cache recovery info",
//				"err", err.Error())
//		}
//
//		fmt.Println("<<<<< AFTER STORING RECOVERY INFO >>>>>")
//	}()
//}

//func ReadRecoveryInfo() *store.KeyRecoveryData {
//	be := Backend()
//	if be == nil {
//		fmt.Println("backend is nil; returning nil")
//		return nil
//	}
//
//	fmt.Println("backend is not nil; returning recovery info")
//
//	retrier := retry.NewExponentialRetrier()
//	typedRetrier := retry.NewTypedRetrier[*store.KeyRecoveryData](retrier)
//
//	ctx, cancel := context.WithTimeout(
//		context.Background(), env.DatabaseOperationTimeout(),
//	)
//	defer cancel()
//
//	cachedRecoveryInfo, err := typedRetrier.RetryWithBackoff(ctx, func() (*store.KeyRecoveryData, error) {
//		return be.LoadKeyRecoveryInfo(ctx)
//	})
//	if err != nil {
//		log.Log().Warn("readRecoveryInfo",
//			"msg", "Failed to load recovery info from cache after retries",
//			"err", err.Error())
//		return nil
//	}
//
//	if cachedRecoveryInfo != nil {
//		fmt.Println("<<<<< RETURNING RECOVERY INFO >>>>>")
//		fmt.Println(cachedRecoveryInfo.RootKey)
//		return cachedRecoveryInfo
//	}
//
//	fmt.Println("<<<<< RETURNING NIL RECOVERY INFO >>>>>")
//	return nil
//}
