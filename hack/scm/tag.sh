#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

VERSION=$(cat app/VERSION.txt)
TAG="v$VERSION"

if git tag -l "$TAG" | grep -q "^$TAG$"; then
	echo "Error: Tag $TAG already exists"
	exit 1
fi

git tag -s "$TAG"
git push origin "$TAG"
