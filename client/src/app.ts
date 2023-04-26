const container = document.getElementById('root');
if (container === null) {
    throw new Error('No matching element found with ID: root');
}

container.innerHTML = '<h1>Hello, World!</h1>' +
    '<p>My name is <strong>Pascal Allen</strong>, and I develop software.</p>';

container.innerHTML += '<h2>Publications</h2>' +
    '<p>' +
    '<a href="https://medium.com/@pascal.allen88/scrum-simplified-880113ed0db" target="_blank">Scrum Simplified</a><br>' +
    'A simple Scrum infrastructure, with insights.' +
    '</p>';

container.innerHTML += '<h2>Go</h2>' +
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
fetch(repoUrl)
    .then(data => {
        return data.json()
    })
    .then(res => {
        res.length > 0 ? container.innerHTML += '<h2>GitHub</h2>' : null;
        for (let i = 0; i < res.length; i++) {
            container.innerHTML += '<p>';
            container.innerHTML += `<a href="${res[i].html_url}" target="_blank">${res[i].name}</a>`;
            container.innerHTML += ` ${new Date(Date.parse(res[i].updated_at)).toLocaleDateString()}<br>`;
            container.innerHTML += `<code>${res[i].description}</code>`;
            container.innerHTML += '</p>';
        }
    });

const registryUrl = "https://registry.npmjs.org/-/v1/search?text=@pascalallen";
fetch(registryUrl)
    .then(data => {
        return data.json()
    })
    .then(res => {
        res.objects.length > 0 ? container.innerHTML += '<h2>NPM</h2>' : null;
        for (let i = 0; i < res.objects.length; i++) {
            container.innerHTML += '<p>';
            container.innerHTML += `<a href="${res.objects[i].package.links.npm}" target="_blank">${res.objects[i].package.name}</a>`;
            container.innerHTML += ` v${res.objects[i].package.version}<br>`;
            container.innerHTML += `<code>${res.objects[i].package.description}</code>`;
            container.innerHTML += '</p>';
        }
    });