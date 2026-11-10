// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {type JSX} from 'react';
import amobiIcon from 'images/amobi-collab.png';

const ProductBrandingFreeEdition = (): JSX.Element => {
    return (
        <span className='globalHeader-leftControls-productBranding-freeEdition'>
            <img
                src={amobiIcon}
                width='28'
                height='28'
                alt=''
                style={{marginRight: 8, borderRadius: 6}}
            />
            <span
                className='globalHeader-leftControls-productBranding-freeEdition-logo'
                aria-label='Amobi Collab'
                style={{fontSize: 18, fontWeight: 600, whiteSpace: 'nowrap'}}
            >
                {'Amobi Collab'}
            </span>
        </span>
    );
};

export default ProductBrandingFreeEdition;
