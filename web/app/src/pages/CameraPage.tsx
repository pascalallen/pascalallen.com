import React, { ReactElement } from 'react';
import { Helmet } from 'react-helmet-async';

const CameraPage = (): ReactElement => {
  return (
    <div className="camera-page">
      <Helmet>
        <title>Pascal Allen - Camera</title>
        <meta name="description" content="Live camera feed" />
      </Helmet>
      <div style={{ width: '100%', maxWidth: '100%' }}>
        <img
          src="/api/v1/camera/stream"
          alt="Live camera feed"
          style={{ width: '100%', height: 'auto', display: 'block' }}
        />
      </div>
    </div>
  );
};

export default CameraPage;
