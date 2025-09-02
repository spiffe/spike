#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# #region SPIRE Server state reset
mkdir -p ./.data
cd ./.data || exit
# shellcheck disable=SC2035
rm -rf *
# #endregion

# #region SPIKE state reset.
DATA_PATH="$HOME"/.spike/data
mkdir -p "$DATA_PATH"
cd "$DATA_PATH" || exit 1
# shellcheck disable=SC2035
rm -rf *
# #endregion
