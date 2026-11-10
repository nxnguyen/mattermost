// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

const documentPreviewMaxBytes = 500000

// The response sandbox also protects direct navigation to the preview URL.
// Omitting allow-same-origin prevents access to app cookies, DOM and preload APIs.
const documentPreviewPolicy = "sandbox allow-scripts; frame-ancestors 'self'; default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data: blob:; font-src data:; media-src data: blob:; connect-src 'none'; frame-src 'none'; object-src 'none'; form-action 'none'; base-uri 'none'"

func sandboxedHTML(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, documentPreviewMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > documentPreviewMaxBytes {
		return nil, errors.New("document exceeds preview size limit")
	}
	document, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var removeNavigationDirectives func(*html.Node)
	removeNavigationDirectives = func(parent *html.Node) {
		for child := parent.FirstChild; child != nil; {
			next := child.NextSibling
			remove := child.Type == html.ElementNode && child.Data == "base"
			if child.Type == html.ElementNode && child.Data == "meta" {
				for _, attr := range child.Attr {
					if strings.EqualFold(attr.Key, "http-equiv") {
						remove = true
					}
				}
			}
			if remove {
				parent.RemoveChild(child)
			} else {
				removeNavigationDirectives(child)
			}
			child = next
		}
	}
	removeNavigationDirectives(document)
	var output bytes.Buffer
	if err := html.Render(&output, document); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
