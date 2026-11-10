// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import amobiIcon from 'images/amobi-collab.png';

export default function AmobiLogo(props: React.HTMLAttributes<HTMLSpanElement>) {
    return (
        <span {...props}>
            <svg viewBox='0 0 500 500' role='img' aria-label='Amobi Collab'>
                <image href={amobiIcon} width='500' height='500'/>
            </svg>
        </span>
    );
}
