// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {fireEvent, renderWithContext, screen} from 'tests/react_testing_utils';
import Constants from 'utils/constants';
import {TestHelper} from 'utils/test_helper';

import DocumentPreview, {isDocumentPreviewFile, documentPreviewURL} from './document_preview';

describe('DocumentPreview', () => {
    const originalFetch = global.fetch;
    const fileUrl = '/api/v4/files/document';
    const fileInfo = TestHelper.getFileInfoMock({id: 'document', name: 'notes.md', extension: 'md', size: 100});

    beforeEach(() => {
        global.fetch = jest.fn().mockResolvedValue({ok: true, text: async () => '# Document\n\n**Hello**'});
    });

    afterEach(() => {
        global.fetch = originalFetch;
    });

    test('supports Markdown and HTML extensions, including uppercase', () => {
        for (const extension of ['md', 'markdown', 'html', 'htm', 'HTML']) {
            expect(isDocumentPreviewFile({extension})).toBe(true);
        }
        expect(isDocumentPreviewFile({extension: 'js'})).toBe(false);
    });

    test('renders Markdown and can switch to escaped source', async () => {
        const getContent = jest.fn();
        const {container} = renderWithContext(
            <DocumentPreview
                fileInfo={fileInfo}
                fileUrl={fileUrl}
                getContent={getContent}
            />,
        );
        expect(await screen.findByRole('heading', {name: 'Document'})).toBeVisible();
        expect(container.querySelector('strong')).toHaveTextContent('Hello');
        expect(getContent).toHaveBeenCalledWith('# Document\n\n**Hello**');
        fireEvent.click(screen.getByRole('button', {name: 'Source'}));
        expect(container.querySelector('pre code')).toHaveTextContent('# Document');
        expect(screen.queryByRole('heading', {name: 'Document'})).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', {name: 'Preview'}));
        expect(screen.getByRole('heading', {name: 'Document'})).toBeVisible();
    });

    test('keeps HTML scripts inside an opaque sandbox', async () => {
        jest.mocked(global.fetch).mockResolvedValue({ok: true, text: async () => '<h1>Example</h1><script>parent.document.title="changed"</script>'} as Response);
        const {container} = renderWithContext(
            <DocumentPreview
                fileInfo={{...fileInfo, name: 'example.html', extension: 'html'}}
                fileUrl={fileUrl}
            />,
        );
        const frame = await screen.findByTitle('example.html');
        expect(frame).toHaveAttribute('sandbox', 'allow-scripts');
        expect(frame).toHaveAttribute('referrerpolicy', 'no-referrer');
        expect(frame.getAttribute('src')).toContain('/api/v4/files/document?amobi_preview=1');
        expect(frame).not.toHaveAttribute('srcdoc');
        expect(container.querySelector('script')).toBeNull();
        fireEvent.click(screen.getByRole('button', {name: 'Source'}));
        expect(container.querySelector('iframe')).toBeNull();
        expect(container.querySelector('script')).toBeNull();
    });

    test('preserves file URL parameters when choosing the sandboxed response', () => {
        const url = new URL(documentPreviewURL('/api/v4/files/document?download=false'));
        expect(url.origin).toBe(window.location.origin);
        expect(url.searchParams.get('download')).toBe('false');
        expect(url.searchParams.get('amobi_preview')).toBe('1');
    });

    test('does not fetch oversized attachments', async () => {
        renderWithContext(
            <DocumentPreview
                fileInfo={{...fileInfo, size: Constants.CODE_PREVIEW_MAX_FILE_SIZE + 1}}
                fileUrl={fileUrl}
            />,
        );
        expect(await screen.findByRole('status')).toHaveTextContent('could not be previewed');
        expect(global.fetch).not.toHaveBeenCalled();
    });

    test('does not fetch external URLs', async () => {
        renderWithContext(
            <DocumentPreview
                fileInfo={fileInfo}
                fileUrl='https://example.com/file.md'
            />,
        );
        expect(await screen.findByRole('status')).toHaveTextContent('could not be previewed');
        expect(global.fetch).not.toHaveBeenCalled();
    });

    test('handles permission errors without rendering a document', async () => {
        jest.mocked(global.fetch).mockResolvedValue({ok: false, status: 403} as Response);
        const {container} = renderWithContext(
            <DocumentPreview
                fileInfo={fileInfo}
                fileUrl={fileUrl}
            />,
        );
        expect(await screen.findByRole('status')).toHaveTextContent('could not be previewed');
        expect(container.querySelector('iframe, article')).toBeNull();
    });
});
