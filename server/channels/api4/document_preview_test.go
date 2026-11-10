// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSandboxedHTML(t *testing.T) {
	output, err := sandboxedHTML(strings.NewReader(`<html lang="vi"><head><base href="https://example.com"><meta HTTP-EQUIV="refresh" content="0;url=https://example.com"></head><body class="demo"><button onclick="this.textContent='OK'">Run</button><script>console.log('demo')</script></body></html>`))
	require.NoError(t, err)
	require.NotContains(t, string(output), "<base")
	require.NotContains(t, string(output), "HTTP-EQUIV")
	require.Contains(t, string(output), `lang="vi"`)
	require.Contains(t, string(output), `class="demo"`)
	require.Contains(t, string(output), "<script>console.log('demo')</script>")
	require.Contains(t, documentPreviewPolicy, "sandbox allow-scripts;")
	require.NotContains(t, documentPreviewPolicy, "allow-same-origin")
	require.Contains(t, documentPreviewPolicy, "connect-src 'none'")
}

func TestSandboxedHTMLSizeLimit(t *testing.T) {
	_, err := sandboxedHTML(strings.NewReader(strings.Repeat("x", documentPreviewMaxBytes+1)))
	require.Error(t, err)
}
