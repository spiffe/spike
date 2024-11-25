//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

type SpikeNexusApiAction string

const KeyApiAction = "action"

type ApiUrl string

const SpikeNexusUrlSecrets ApiUrl = "/v1/store/secrets"
const SpikeNexusUrlLogin ApiUrl = "/v1/auth/login"
const SpikeNexusUrlInit ApiUrl = "/v1/auth/init"

const SpikeNexusUrlPolicy ApiUrl = "/v1/acl/policy"

const SpikeKeeperUrlKeep ApiUrl = "/v1/store/keep"

const ActionNexusAdminLogin SpikeNexusApiAction = "admin-login"
const ActionNexusCheck SpikeNexusApiAction = "check"
const ActionNexusGet SpikeNexusApiAction = "get"
const ActionNexusDelete SpikeNexusApiAction = "delete"
const ActionNexusUndelete SpikeNexusApiAction = "undelete"
const ActionNexusList SpikeNexusApiAction = "list"
const ActionNexusDefault SpikeNexusApiAction = ""

type SpikeKeeperApiAction string

const ActionKeeperRead SpikeKeeperApiAction = "read"
const ActionKeeperDefault SpikeKeeperApiAction = ""
