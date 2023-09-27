import React, { MouseEvent, ReactElement, useEffect, useRef, useState } from 'react';
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
  }, []);

  useEffect(() => {
    fetch(npmUrl)
      .then(response => response.json())
      .then(data => setPackages(data.objects));
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
    <div className="index-page">
      <Helmet>
        <title>Pascal Allen - Home</title>
        <meta name="description" content="Welcome to the home page for pascalallen.com" />
      </Helmet>
      <div className="navbar-container">
        <div id="navbar" className="navbar">
          <div id="section-links" className="section-links">
            <a href="#technology" onClick={scrollToLocation}>
              Technology
            </a>
            <a href="#publications" onClick={scrollToLocation}>
              Publications
            </a>
            <a href="#golang" onClick={scrollToLocation}>
              Golang
            </a>
            <a href="#github" onClick={scrollToLocation}>
              GitHub
            </a>
            <a href="#npm" onClick={scrollToLocation}>
              NPM
            </a>
          </div>
          <div id="social-links" className="social-links">
            <a href="https://www.linkedin.com/in/pascal-allen-942749112/" target="_blank" rel="noreferrer">
              <i className="fa-brands fa-linkedin" />
            </a>
            <a href="https://github.com/pascalallen" target="_blank" rel="noreferrer">
              <i className="fa-brands fa-github" />
            </a>
          </div>
        </div>
        <div className="hamburger">
          <a onClick={handleToggleNav}>
            <i className="fa-solid fa-bars" />
          </a>
        </div>
      </div>
      <header className="header">
        <div>
          <h1>Pascal Allen</h1>
          <p>Software Developer</p>
        </div>
      </header>
      <section id="technology" className="technology-section">
        <div>
          <h2>
            This Site Runs On{' '}
            <a href="#technology" onClick={scrollToLocation}>
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
            <a href="#publications" onClick={scrollToLocation}>
              #
            </a>
          </h2>
          <p>
            <a
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
              href="https://pascalallen.medium.com/dispatching-events-with-react-and-typescript-89f80f07635f"
              target="_blank"
              rel="noreferrer">
              Dispatching Events With React and TypeScript
            </a>
            <br />A demonstration on how to dispatch and listen to events with React and TypeScript.
          </p>
          <p>
            <a
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
            <a href="https://pascalallen.medium.com/scrum-simplified-880113ed0db" target="_blank" rel="noreferrer">
              Scrum Simplified
            </a>
            <br />A simple Scrum infrastructure, with insights.
          </p>
          <p>
            <a
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
            <a href="#golang" onClick={scrollToLocation}>
              #
            </a>
          </h2>
          <p>
            <a href="https://pkg.go.dev/github.com/pascalallen/pubsub" target="_blank" rel="noreferrer">
              pubsub
            </a>{' '}
            v1.0.0
            <br />
            <code>
              pubsub is a Go module that offers a concurrent pub/sub service leveraging goroutines and channels.
            </code>
          </p>
          <p>
            <a href="https://pkg.go.dev/github.com/pascalallen/hmac" target="_blank" rel="noreferrer">
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
              <a href="#github" onClick={scrollToLocation}>
                #
              </a>
            </h2>
            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            {repos.map((repo: any, index: number) => (
              <p key={`repo-${index}`}>
                <a href={repo.html_url} target="_blank" rel="noreferrer">
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
              <a href="#npm" onClick={scrollToLocation}>
                #
              </a>
            </h2>
            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            {packages.map((pkg: any, index: number) => (
              <p key={`pkg-${index}`}>
                <a href={pkg.package.links.npm} target="_blank" rel="noreferrer">
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
