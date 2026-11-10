// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {renderWithContext, screen} from 'tests/react_testing_utils';

import Header from './header';

describe('Amobi login header', () => {
    it('shows a single Amobi identity without the upstream edition banner', () => {
        renderWithContext(<Header/>, {
            entities: {
                general: {
                    config: {SiteName: 'Amobi Collab'},
                    license: {IsLicensed: 'false'},
                },
            },
        });
        expect(screen.getAllByRole('link', {name: 'Amobi Collab'})).toHaveLength(1);
        expect(screen.getByRole('img', {name: 'Amobi Collab'})).toBeVisible();
        expect(screen.queryByText(/TEAM EDITION/)).not.toBeInTheDocument();
        expect(screen.queryByText(/Mattermost/)).not.toBeInTheDocument();
    });
});
