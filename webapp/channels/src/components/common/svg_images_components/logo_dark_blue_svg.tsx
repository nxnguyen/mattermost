// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import amobiIcon from 'images/amobi-collab.png';

type Props = {width?: number; height?: number; className?: string};

export default function AmobiLogo(props: Props) {
    return (
        <svg
            className={props.className}
            width={props.width || 206}
            height={props.height || 36}
            viewBox='0 0 206 36'
            role='img'
            aria-label='Amobi Collab'
        >
            <image href={amobiIcon} width='36' height='36'/>
            <text x='46' y='25' fontSize='22' fontWeight='600' fontFamily='Open Sans, sans-serif' fill='currentColor'>
                {'Amobi Collab'}
            </text>
        </svg>
    );
}
