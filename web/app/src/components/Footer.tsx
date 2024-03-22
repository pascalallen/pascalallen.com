import React, { ReactElement } from 'react';

const Footer = (): ReactElement => {
  return (
    <footer id="footer" className="footer">
      <p className="copyright-desktop">© 2024 Pascal Allen & Crimson Drive Design LLC</p>
      <p className="copyright-mobile">© 2024 Pascal Allen & CDD LLC</p>
      <a href="https://www.termsfeed.com/live/acb0c6cd-2718-465b-9339-7997e07f7ca9" target="_blank" rel="noreferrer">
        Terms
      </a>
      <a
        href="https://www.privacypolicies.com/live/47848157-ff99-4a14-a488-f5e0fb1b89af"
        target="_blank"
        rel="noreferrer">
        Privacy
      </a>
    </footer>
  );
};

export default Footer;
