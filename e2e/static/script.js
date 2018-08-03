(async () => {
    const rawResponse = await fetch('/', {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({a: 1, b: 'Textual content'})
    });
    const content = await rawResponse.json();

    console.log(content);
})();