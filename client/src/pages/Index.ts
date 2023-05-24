export default async function Index(): Promise<string> {
    let content = '<h1>Hello, World!</h1>' +
        '<p>My name is <strong>Pascal Allen</strong>, and I develop software.</p>';

    content += '<h2>Publications</h2>' +
        '<p>' +
        '<a href="https://medium.com/@pascal.allen88/scrum-simplified-880113ed0db" target="_blank">Scrum Simplified</a><br>' +
        'A simple Scrum infrastructure, with insights.' +
        '</p>' +
        '<p>' +
        '<a href="https://www.bizjournals.com/sanantonio/news/2016/11/23/divergent-career-paths-how-tech-talent-is-leaking.html" target="_blank">Divergent Career Paths</a><br>' +
        'San Antonio Business Journal: How tech talent is leaking out of San Antonio.' +
        '</p>';

    content += '<h2>Go</h2>' +
        '<p>' +
        '<a href="https://pkg.go.dev/github.com/pascalallen/pubsub" target="_blank">pubsub</a>' +
        ' v1.0.0<br>' +
        '<code>pubsub is a Go module that offers a concurrent pub/sub service leveraging goroutines and channels.</code>' +
        '</p>' +
        '<p>' +
        '<a href="https://pkg.go.dev/github.com/pascalallen/hmac" target="_blank">hmac</a>' +
        ' v1.0.1<br>' +
        '<code>hmac is a Go module that offers services for HTTP HMAC authentication.</code>' +
        '</p>';

    const repoUrl = "https://api.github.com/users/pascalallen/repos?per_page=10&sort=updated_at&direction=desc";
    try {
        const response = await fetch(repoUrl);
        const data = await response.json();
        data.length > 0 ? content += '<h2>GitHub</h2>' : null;
        for (let i = 0; i < data.length; i++) {
            content += '<p>';
            content += `<a href="${data[i].html_url}" target="_blank">${data[i].name}</a>`;
            content += ` ${new Date(Date.parse(data[i].updated_at)).toLocaleDateString()}<br>`;
            content += `<code>${data[i].description}</code>`;
            content += '</p>';
        }
    } catch (err) {
        console.error(err);
    }

    const registryUrl = "https://registry.npmjs.org/-/v1/search?text=@pascalallen";
    try {
        const response = await fetch(registryUrl);
        const data = await response.json();
        data.objects.length > 0 ? content += '<h2>NPM</h2>' : null;
        for (let i = 0; i < data.objects.length; i++) {
            content += '<p>';
            content += `<a href="${data.objects[i].package.links.npm}" target="_blank">${data.objects[i].package.name}</a>`;
            content += ` v${data.objects[i].package.version}<br>`;
            content += `<code>${data.objects[i].package.description}</code>`;
            content += '</p>';
        }
    } catch (err) {
        console.error(err);
    }

    return content;
};
