//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

type SpikeNexusApiAction string

const ActionAdminLogin SpikeNexusApiAction = "admin-login"
const ActionCheck SpikeNexusApiAction = "check"
const ActionGet SpikeNexusApiAction = "get"
const ActionDelete SpikeNexusApiAction = "delete"
const ActionUndelete SpikeNexusApiAction = "undelete"
const ActionList SpikeNexusApiAction = "list"
const ActionDefault SpikeNexusApiAction = ""
