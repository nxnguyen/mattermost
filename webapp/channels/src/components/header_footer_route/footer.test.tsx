// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {renderWithContext, screen} from 'tests/react_testing_utils';

import Footer from './footer';

describe('Amobi login footer', () => {
    it('shows the Amobi name and configured support links', () => {
        renderWithContext(<Footer/>, {
            entities: {
                general: {
                    config: {
                        AboutLink: 'https://152.42.230.207.sslip.io/static/amobi-open-source-notices.txt',
                        HelpLink: 'https://152.42.230.207.sslip.io/static/amobi-help.html',
                        TermsOfServiceLink: 'https://mattermost.com/pl/terms-of-use/',
                        PrivacyPolicyLink: '',
                    },
                },
            },
        });

        expect(screen.getByText(`© ${new Date().getFullYear()} Amobi Collab`)).toBeVisible();
        expect(screen.getByRole('link', {name: 'About'})).toHaveAttribute('href', expect.stringContaining('amobi-open-source-notices.txt'));
        expect(screen.getByRole('link', {name: 'Help'})).toHaveAttribute('href', expect.stringContaining('amobi-help.html'));
        expect(screen.queryByRole('link', {name: 'Privacy Policy'})).not.toBeInTheDocument();
        expect(screen.queryByRole('link', {name: 'Terms'})).not.toBeInTheDocument();
    });
    it('keeps an explicitly configured Amobi terms link', () => {
        renderWithContext(<Footer/>, {
            entities: {general: {config: {TermsOfServiceLink: 'https://collab.example/terms'}}},
        });
        expect(screen.getByRole('link', {name: 'Terms'})).toHaveAttribute('href', 'https://collab.example/terms');
    });

});
