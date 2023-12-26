import React, { MouseEvent, ReactElement, useEffect, useRef, useState } from 'react';
import axios from 'axios';
import { Helmet } from 'react-helmet-async';
import { useLocation } from 'react-router';
import env, { EnvKey } from '@utilities/env';
import DockerLogo from '@assets/images/docker-logo.svg';
import GoLogo from '@assets/images/go-logo.svg';
import K8sLogo from '@assets/images/k8s-logo.svg';
import NginxLogo from '@assets/images/nginx-logo.svg';
import PostgresLogo from '@assets/images/postgres-logo.svg';
import ReactLogo from '@assets/images/react-logo.svg';
import SassLogo from '@assets/images/sass-logo.svg';
import TsLogo from '@assets/images/ts-logo.svg';
import UbuntuLogo from '@assets/images/ubuntu-logo.svg';
import WebAssemblyLogo from '@assets/images/webassembly-logo.svg';
import WebpackLogo from '@assets/images/webpack-logo.svg';
import Footer from '@components/Footer';

const gitHubUrl = 'https://api.github.com/users/pascalallen/repos?per_page=10&sort=updated_at&direction=desc';
const npmUrl = 'https://registry.npmjs.org/-/v1/search?text=@pascalallen';

type State = {
  repos: [];
  packages: [];
};

const initialState: State = {
  repos: [],
  packages: []
};

const IndexPage = (): ReactElement => {
  let { hash } = useLocation();
  const scrolledRef = useRef(false);
  const hashRef = useRef(hash);

  const [repos, setRepos] = useState(initialState.repos);
  const [packages, setPackages] = useState(initialState.packages);
  const navbar: HTMLElement | null = document.getElementById('navbar');
  const sectionLinks: HTMLElement | null = document.getElementById('section-links');
  const socialLinks: HTMLElement | null = document.getElementById('social-links');

  useEffect(() => {
    fetch(gitHubUrl, { headers: { Authorization: `token ${env(EnvKey.GITHUB_TOKEN)}` } })
      .then(response => response.json())
      .then(data => setRepos(data));

    fetch(npmUrl)
      .then(response => response.json())
      .then(data => setPackages(data.objects));

    if (`${env(EnvKey.APP_ENV)}` === 'prod' || `${env(EnvKey.APP_ENV)}` === 'production') {
      const user = {
        language: window.navigator.language,
        user_agent: window.navigator.userAgent
      };
      axios.post(`${env(EnvKey.SLACK_DM_URL)}`, JSON.stringify({ text: JSON.stringify(user, null, 4) }));
    }
  }, []);

  const scrollToLocation = (event?: MouseEvent) => {
    if (event !== undefined) {
      event.preventDefault();
      hash = event.currentTarget.getAttribute('href') ?? '';
      history.replaceState(null, '', hash);
    }

    if (hash) {
      if (hashRef.current !== hash) {
        hashRef.current = hash;
        scrolledRef.current = false;
      }

      if (!scrolledRef.current) {
        const id = hash.replace('#', '');
        const element = document.getElementById(id);
        if (element) {
          element.scrollIntoView({ behavior: 'smooth' });
          scrolledRef.current = true;
          navbar?.classList.remove('navbar-mobile-open');
          sectionLinks?.classList.remove('d-flex');
          socialLinks?.classList.remove('d-flex');
        }
      }
    }
  };

  useEffect(scrollToLocation);

  const handleToggleNav = (event: MouseEvent): void => {
    event.preventDefault();

    !navbar?.classList.contains('navbar-mobile-open')
      ? navbar?.classList.add('navbar-mobile-open')
      : navbar?.classList.contains('navbar-mobile-open')
        ? navbar?.classList.remove('navbar-mobile-open')
        : null;
    !sectionLinks?.classList.contains('d-flex')
      ? sectionLinks?.classList.add('d-flex')
      : sectionLinks?.classList.contains('d-flex')
        ? sectionLinks?.classList.remove('d-flex')
        : null;
    !socialLinks?.classList.contains('d-flex')
      ? socialLinks?.classList.add('d-flex')
      : socialLinks?.classList.contains('d-flex')
        ? socialLinks?.classList.remove('d-flex')
        : null;
  };

  return (
    <div id="index-page" className="index-page">
      <Helmet>
        <title>Pascal Allen - Home</title>
        <meta name="description" content="Welcome to the home page for pascalallen.com" />
      </Helmet>
      <div className="navbar-container">
        <div id="navbar" className="navbar">
          <div id="section-links" className="section-links">
            <a id="navbar-technology-link" href="#technology" onClick={scrollToLocation}>
              Technology
            </a>
            <a id="navbar-publications-link" href="#publications" onClick={scrollToLocation}>
              Publications
            </a>
            <a id="navbar-golang-link" href="#golang" onClick={scrollToLocation}>
              Golang
            </a>
            <a id="navbar-github-link" href="#github" onClick={scrollToLocation}>
              GitHub
            </a>
            <a id="navbar-npm-link" href="#npm" onClick={scrollToLocation}>
              NPM
            </a>
          </div>
          <div id="social-links" className="social-links">
            <a
              id="social-linkedin-link"
              href="https://www.linkedin.com/in/pascal-allen-942749112/"
              target="_blank"
              rel="noreferrer">
              <i className="fa-brands fa-linkedin" />
            </a>
            <a id="social-github-link" href="https://github.com/pascalallen" target="_blank" rel="noreferrer">
              <i className="fa-brands fa-github" />
            </a>
          </div>
        </div>
        <div className="hamburger">
          <a id="hamburger-link" onClick={handleToggleNav}>
            <i className="fa-solid fa-bars" />
          </a>
        </div>
      </div>
      <header className="header">
        <div>
          <h1>Pascal Allen</h1>
          <p className="profession">
            <span id="profession-text" />
            <span className="blink">_</span>
          </p>
        </div>
      </header>
      <section id="technology" className="technology-section">
        <div>
          <h2>
            This Site Runs On{' '}
            <a id="technology-hashtag" href="#technology" onClick={scrollToLocation}>
              #
            </a>
          </h2>
          <div className="technology-list">
            <img className="tech-image" src={UbuntuLogo} alt="Ubuntu logo" title="Ubuntu" width={60} />
            <img className="tech-image" src={NginxLogo} alt="Nginx logo" title="Nginx" width={60} />
            <img className="tech-image" src={K8sLogo} alt="Kubernetes logo" title="Kubernetes" width={60} />
            <img className="tech-image" src={DockerLogo} alt="Docker logo" title="Docker" width={60} />
            <img className="tech-image" src={PostgresLogo} alt="Postgres logo" title="Postgres" width={60} />
            <img className="tech-image" src={GoLogo} alt="Go logo" title="Go" width={60} />
            <img className="tech-image" src={WebAssemblyLogo} alt="WebAssembly logo" title="WebAssembly" width={60} />
            <img className="tech-image" src={ReactLogo} alt="React logo" title="React" width={60} />
            <img className="tech-image" src={TsLogo} alt="TypeScript logo" title="TypeScript" width={60} />
            <img className="tech-image" src={SassLogo} alt="Sass logo" title="Sass" width={60} />
            <img className="tech-image" src={WebpackLogo} alt="Webpack logo" title="Webpack" width={60} />
          </div>
        </div>
      </section>
      <section id="publications" className="publications-section">
        <div id="background-circles" className="background-circles">
          <div />
          <div />
          <div />
          <div />
          <div />
          <div />
          <div />
          <div />
          <div />
        </div>
        <div>
          <h2>
            Publications{' '}
            <a id="publications-hashtag" href="#publications" onClick={scrollToLocation}>
              #
            </a>
          </h2>
          <p>
            <a
              id="medium-cron-link"
              href="https://pascalallen.medium.com/automate-your-deployments-with-cron-7174ecb9f52f"
              target="_blank"
              rel="noreferrer">
              Automate Your Deployments With Cron
            </a>
            <br />A basic set of instructions to automatically deploy your app with cron.
          </p>
          <p>
            <a
              id="medium-grpc-link"
              href="https://pascalallen.medium.com/how-to-build-a-grpc-server-in-go-943f337c4e05"
              target="_blank"
              rel="noreferrer">
              How To: Build a gRPC Server In Go
            </a>
            <br />
            Learn how to build a gRPC server and client in Go.
          </p>
          <p>
            <a
              id="medium-framework-link"
              href="https://pascalallen.medium.com/developing-a-framework-for-any-project-9cf7dac82ffe"
              target="_blank"
              rel="noreferrer">
              Developing a Framework for Any Project
            </a>
            <br />A resource for designing and developing a product that can be easily maintained and extended by future
            software developers and domain experts.
          </p>
          <p>
            <a
              id="medium-sse-go-link"
              href="https://pascalallen.medium.com/streaming-server-sent-events-with-go-8cc1f615d561"
              target="_blank"
              rel="noreferrer">
              Streaming Server-Sent Events With Go
            </a>
            <br />
            This publication demonstrates how to stream server-sent events over HTTP with Go.
          </p>
          <p>
            <a
              id="medium-event-dispatch-react-link"
              href="https://pascalallen.medium.com/dispatching-events-with-react-and-typescript-89f80f07635f"
              target="_blank"
              rel="noreferrer">
              Dispatching Events With React and TypeScript
            </a>
            <br />A demonstration on how to dispatch and listen to events with React and TypeScript.
          </p>
          <p>
            <a
              id="medium-jwt-go-link"
              href="https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8"
              target="_blank"
              rel="noreferrer">
              JWT Authentication With Go
            </a>
            <br />A walk-through of creating, validating, and refreshing JSON Web Tokens using the HMAC signing method
            with Go.
          </p>
          <p>
            <a
              id="medium-wasm-go-link"
              href="https://pascalallen.medium.com/how-to-compile-a-webassembly-module-from-go-a9ed5f831582"
              target="_blank"
              rel="noreferrer">
              How To: Compile a WebAssembly Module From Go
            </a>
            <br />
            Learn how to compile a WebAssembly module from Go.
          </p>
          <p>
            <a
              id="medium-deploy-k8s-link"
              href="https://pascalallen.medium.com/how-to-deploy-to-kubernetes-76c42e5ea28c"
              target="_blank"
              rel="noreferrer">
              How To: Deploy to Kubernetes
            </a>
            <br />
            Learn how to deploy to Kubernetes.
          </p>
          <p>
            <a
              id="medium-docker-go-link"
              href="https://pascalallen.medium.com/how-to-build-a-containerized-web-app-in-go-73f42619a193"
              target="_blank"
              rel="noreferrer">
              How To: Build a Containerized Web App In Go
            </a>
            <br />
            Learn how to build a containerized web app with Docker and Go.
          </p>
          <p>
            <a
              id="medium-npm-package-link"
              href="https://pascalallen.medium.com/releasing-packages-to-github-and-the-npm-registry-8ff6c3789bc8"
              target="_blank"
              rel="noreferrer">
              Releasing Packages to GitHub and the npm Registry
            </a>
            <br />
            This publication describes a simple process I follow to tag and release a new package version to GitHub and
            the npm Registry.
          </p>
          <p>
            <a
              id="medium-scrum-link"
              href="https://pascalallen.medium.com/scrum-simplified-880113ed0db"
              target="_blank"
              rel="noreferrer">
              Scrum Simplified
            </a>
            <br />A simple Scrum infrastructure, with insights.
          </p>
          <p>
            <a
              id="medium-sabj-link"
              href="https://www.bizjournals.com/sanantonio/news/2016/11/23/divergent-career-paths-how-tech-talent-is-leaking.html"
              target="_blank"
              rel="noreferrer">
              Divergent Career Paths
            </a>
            <br />
            San Antonio Business Journal: How tech talent is leaking out of San Antonio.
          </p>
        </div>
      </section>
      <section id="golang" className="go-section">
        <div>
          <h2>
            Go{' '}
            <a id="golang-hashtag" href="#golang" onClick={scrollToLocation}>
              #
            </a>
          </h2>
          <p>
            <a
              id="pubsub-package-link"
              href="https://pkg.go.dev/github.com/pascalallen/pubsub"
              target="_blank"
              rel="noreferrer">
              pubsub
            </a>{' '}
            v1.0.0
            <br />
            <code>
              pubsub is a Go module that offers a concurrent pub/sub service leveraging goroutines and channels.
            </code>
          </p>
          <p>
            <a
              id="hmac-package-link"
              href="https://pkg.go.dev/github.com/pascalallen/hmac"
              target="_blank"
              rel="noreferrer">
              hmac
            </a>{' '}
            v1.0.1
            <br />
            <code>hmac is a Go module that offers services for HTTP HMAC authentication.</code>
          </p>
        </div>
      </section>
      {repos.length > 0 && (
        <section id="github" className="github-section">
          <div id="background-rectangles" className="background-rectangles">
            <div />
            <div />
            <div />
            <div />
            <div />
            <div />
            <div />
            <div />
            <div />
          </div>
          <div>
            <h2>
              GitHub{' '}
              <a id="github-hashtag" href="#github" onClick={scrollToLocation}>
                #
              </a>
            </h2>
            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            {repos.map((repo: any, index: number) => (
              <p key={`repo-${index}`}>
                <a id={`${repo.name}-repo-link`} href={repo.html_url} target="_blank" rel="noreferrer">
                  {repo.name}
                </a>{' '}
                {new Date(Date.parse(repo.updated_at)).toLocaleDateString()}
                <br />
                <code>{repo.description}</code>
              </p>
            ))}
          </div>
        </section>
      )}
      {packages.length > 0 && (
        <section id="npm" className="npm-section">
          <div>
            <h2>
              NPM{' '}
              <a id="npm-hashtag" href="#npm" onClick={scrollToLocation}>
                #
              </a>
            </h2>
            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            {packages.map((pkg: any, index: number) => (
              <p key={`pkg-${index}`}>
                <a id={`${pkg.package.name}-npm-link`} href={pkg.package.links.npm} target="_blank" rel="noreferrer">
                  {pkg.package.name}
                </a>{' '}
                v{pkg.package.version}
                <br />
                <code>{pkg.package.description}</code>
              </p>
            ))}
          </div>
        </section>
      )}
      <Footer />
    </div>
  );
};

export default IndexPage;
