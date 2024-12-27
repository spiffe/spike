//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

type SpikeNexusApiAction string

const KeyApiAction = "action"

const ActionNexusGet SpikeNexusApiAction = "get"
const ActionNexusDelete SpikeNexusApiAction = "delete"
const ActionNexusUndelete SpikeNexusApiAction = "undelete"
const ActionNexusList SpikeNexusApiAction = "list"
const ActionNexusDefault SpikeNexusApiAction = ""

type SpikeKeeperApiAction string

const ActionKeeperDefault SpikeKeeperApiAction = ""
