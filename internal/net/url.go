//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

type ApiUrl string

const SpikeNexusUrlSecrets ApiUrl = "/v1/store/secrets"
const SpikeNexusUrlSecretsMetadata ApiUrl = "/v1/store/secrets/metadata"
const SpikeNexusUrlPolicy ApiUrl = "/v1/acl/policy"

const SpikeNexusUrlOperatorRestore = "/v1/operator/restore"
const SpikeNexusUrlOperatorRecover = "/v1/operator/recover"

const SpikeKeeperUrlKeep ApiUrl = "/v1/store/keep"

const SpikeKeeperUrlContribute ApiUrl = "/v1/store/contribute"
const SpikeKeeperUrlShard ApiUrl = "/v1/store/shard"
