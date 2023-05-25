import React, {ReactElement, useEffect, useState} from "react";

const gitHubUrl = "https://api.github.com/users/pascalallen/repos?per_page=10&sort=updated_at&direction=desc";
const npmUrl = "https://registry.npmjs.org/-/v1/search?text=@pascalallen";

type State = {
    repos: [];
    packages: [];
};

const initialState: State = {
    repos: [],
    packages: []
};

const IndexPage = (): ReactElement => {
    const [repos, setRepos] = useState(initialState.repos);
    const [packages, setPackages] = useState(initialState.packages);

    useEffect(() => {
        fetch(gitHubUrl)
            .then(response => response.json())
            .then(data => setRepos(data));
    }, []);

    useEffect(() => {
        fetch(npmUrl)
            .then(response => response.json())
            .then(data => setPackages(data.objects));
    }, []);

    return (
        <div className="index-page">
            <header className="header">
                <h1>Hello, World!</h1>
                <p>My name is <strong>Pascal Allen</strong>, and I develop software.</p>
            </header>
            <section className="publications-section">
                <h2>Publications</h2>
                <p>
                    <a href="https://medium.com/@pascal.allen88/scrum-simplified-880113ed0db" target="_blank">Scrum Simplified</a>
                    <br/>
                    A simple Scrum infrastructure, with insights.
                </p>
                <p>
                    <a href="https://www.bizjournals.com/sanantonio/news/2016/11/23/divergent-career-paths-how-tech-talent-is-leaking.html" target="_blank">Divergent Career Paths</a>
                    <br/>
                    San Antonio Business Journal: How tech talent is leaking out of San Antonio.
                </p>
            </section>
            <section className="go-section">
                <h2>Go</h2>
                <p>
                    <a href="https://pkg.go.dev/github.com/pascalallen/pubsub" target="_blank">pubsub</a>
                    v1.0.0
                    <br/>
                    <code>pubsub is a Go module that offers a concurrent pub/sub service leveraging goroutines and channels.</code>
                </p>
                <p>
                    <a href="https://pkg.go.dev/github.com/pascalallen/hmac" target="_blank">hmac</a>
                    v1.0.1
                    <br/>
                    <code>hmac is a Go module that offers services for HTTP HMAC authentication.</code>
                </p>
            </section>
            {repos.length > 0 &&
                <section>
                    <h2>GitHub</h2>
                    {repos.map((repo: any, index: number) => (
                        <p key={`repo-${index}`}>
                            <a href={repo.html_url} target="_blank">{repo.name}</a>{' '}{new Date(Date.parse(repo.updated_at)).toLocaleDateString()}
                            <br/>
                            <code>{repo.description}</code>
                        </p>
                    ))}
                </section>
            }
            {packages.length > 0 &&
                <section>
                    <h2>NPM</h2>
                    {packages.map((pkg: any, index: number) => (
                        <p key={`pkg-${index}`}>
                            <a href={pkg.package.links.npm} target="_blank">{pkg.package.name}</a>{' '}v{pkg.package.version}
                            <br/>
                            <code>{pkg.package.description}</code>
                        </p>
                    ))}
                </section>
            }
        </div>
    );
};

export default IndexPage;
