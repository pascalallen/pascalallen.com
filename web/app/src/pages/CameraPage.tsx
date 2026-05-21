import React, { ReactElement } from 'react';
import { Helmet } from 'react-helmet-async';

const CameraPage = (): ReactElement => {
  return (
    <div className="camera-page">
      <Helmet>
        <title>Pascal Allen - Camera</title>
        <meta name="description" content="Live camera feed" />
      </Helmet>
      <img src="/api/v1/camera/stream" alt="Live camera feed" />
    </div>
  );
};

export default CameraPage;
