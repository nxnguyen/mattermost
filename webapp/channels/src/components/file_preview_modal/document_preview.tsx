// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useEffect, useState} from 'react';
import {FormattedMessage} from 'react-intl';

import type {FileInfo} from '@mattermost/types/files';

import FileInfoPreview from 'components/file_info_preview';
import Markdown from 'components/markdown';
import LoadingSpinner from 'components/widgets/loading/loading_spinner';

import Constants from 'utils/constants';

import './document_preview.scss';

type Props = {
    fileInfo: FileInfo;
    fileUrl: string;
    getContent?: (content: string) => void;
};

export function isDocumentPreviewFile(fileInfo: Pick<FileInfo, 'extension'>): boolean {
    return ['html', 'htm', 'md', 'markdown'].includes(fileInfo.extension?.toLowerCase());
}

export function documentPreviewURL(fileUrl: string): string {
    // A navigated document gets its own restrictive response CSP. srcDoc would
    // inherit the parent script policy, preventing self-contained HTML demos.
    const url = new URL(fileUrl, window.location.href);
    url.searchParams.set('amobi_preview', '1');
    return url.href;
}

export default function DocumentPreview({fileInfo, fileUrl, getContent}: Props) {
    const [content, setContent] = useState('');
    const [status, setStatus] = useState<'loading' | 'success' | 'fail'>('loading');
    const [showSource, setShowSource] = useState(false);
    const isHtml = ['html', 'htm'].includes(fileInfo.extension.toLowerCase());

    useEffect(() => {
        const controller = new AbortController();
        setStatus('loading');
        setContent('');
        setShowSource(false);

        const load = async () => {
            try {
                if (fileInfo.size > Constants.CODE_PREVIEW_MAX_FILE_SIZE || new URL(fileUrl, window.location.href).origin !== window.location.origin) {
                    throw new Error('Unsupported document preview');
                }
                const response = await fetch(fileUrl, {signal: controller.signal, credentials: 'same-origin'});
                if (!response.ok) {
                    throw new Error('Document unavailable');
                }
                const text = await response.text();
                if (text.length > Constants.CODE_PREVIEW_MAX_FILE_SIZE) {
                    throw new Error('Document too large');
                }
                if (!controller.signal.aborted) {
                    setContent(text);
                    getContent?.(text);
                    setStatus('success');
                }
            } catch {
                if (!controller.signal.aborted) {
                    setStatus('fail');
                }
            }
        };
        load();
        return () => controller.abort();
    }, [fileInfo.size, fileUrl, getContent]);

    if (status === 'loading') {
        return <div className='view-image__loading'><LoadingSpinner/></div>;
    }

    if (status === 'fail') {
        return (
            <div className='document-preview__error'>
                <p role='status'>
                    <FormattedMessage
                        id='amobi.document_preview.unavailable'
                        defaultMessage='This document could not be previewed. Preview supports files up to 500 KB.'
                    />
                </p>
                <FileInfoPreview
                    fileInfo={fileInfo}
                    fileUrl={fileUrl}
                />
            </div>
        );
    }

    return (
        <section
            className='document-preview'
            aria-label={fileInfo.name}
        >
            <div className='document-preview__toolbar'>
                <button
                    type='button'
                    className='btn btn-tertiary'
                    aria-pressed={!showSource}
                    onClick={() => setShowSource(false)}
                >
                    <FormattedMessage
                        id='amobi.document_preview.preview'
                        defaultMessage='Preview'
                    />
                </button>
                <button
                    type='button'
                    className='btn btn-tertiary'
                    aria-pressed={showSource}
                    onClick={() => setShowSource(true)}
                >
                    <FormattedMessage
                        id='amobi.document_preview.source'
                        defaultMessage='Source'
                    />
                </button>
            </div>
            {showSource && <pre className='document-preview__source'><code>{content}</code></pre>}
            {!showSource && isHtml && (
                <iframe
                    className='document-preview__html'
                    title={fileInfo.name}
                    sandbox='allow-scripts'
                    referrerPolicy='no-referrer'
                    src={documentPreviewURL(fileUrl)}
                />
            )}
            {!showSource && !isHtml && (
                <article className='document-preview__markdown markdown__content'>
                    <Markdown
                        message={content}
                        options={{atMentions: false, channelNamesMap: {}}}
                    />
                </article>
            )}
        </section>
    );
}
