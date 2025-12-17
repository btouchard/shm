// SPDX-License-Identifier: AGPL-3.0-or-later

package web

import "embed"

//go:embed index.html assets/*
var Assets embed.FS
