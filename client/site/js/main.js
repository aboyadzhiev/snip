(() => {
    'use strict'

    const form = document.querySelector("#url-shortening-form");
    const url = form.querySelector("#url");

    const resultTable = document.querySelector("#result-table");
    const snipURLInput = document.getElementById("result-snip-url");
    const longURLInput = document.getElementById("result-long-url");
    const snipURLAnchor = document.getElementById("result-anchor-snip-url");


    const shortenURL = async () => {
        const payload = {"url": url.value};

        try {
            const response = await fetch('/api/v1/shortened-url', {
                method: 'POST',
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify(payload),
            });

            if (response.ok) {
                const responseBody = await response.json();
                longURLInput.value = url.value;
                snipURLInput.value = responseBody.shortenURL;
                snipURLAnchor.href = responseBody.shortenURL;

                return true;
            }
        } catch (e) {
            console.error(e);
        }

        return false;
    }

    const appendAlert = (message, type) => {
        const wrapper = document.createElement('div')
        wrapper.innerHTML = [
            `<div class="col-12 text-center alert alert-${type} alert-dismissible" role="alert">`,
            `    <div>${message}</div>`,
            '    <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>',
            '</div>',
        ].join('')

        const container = document.querySelector("#message");
        container.append(wrapper);
    }

    form.addEventListener('submit', (event) => {
        event.preventDefault();
        const submitBtn = form.querySelector("#shorten-url");
        submitBtn.setAttribute('disabled', 'disabled');

        const loadingIndicator = document.createElement('span');
        loadingIndicator.setAttribute('class', 'spinner-border spinner-border-sm mx-1');
        loadingIndicator.setAttribute('aria-hidden', 'true');

        const loadingTxt = document.createElement('span');
        loadingTxt.setAttribute('role', 'status');
        loadingTxt.innerText = 'Shortening...';

        submitBtn.innerHTML = null;
        submitBtn.appendChild(loadingIndicator);
        submitBtn.appendChild(loadingTxt);

        shortenURL().then((success) => {
            if (success) {
                form.classList.add("d-none");

                resultTable.classList.remove("d-none");

                appendAlert("Success! Your URL has been shortened. Thank you using our service!", "success");
            } else {
                form.classList.remove("d-none");
                resultTable.classList.add("d-none");

                appendAlert("Error! An unexpected error occurred while shortening your URL. Please try again later...", "danger");
            }
            submitBtn.removeChild(loadingTxt);
            submitBtn.removeChild(loadingIndicator);
            submitBtn.innerHTML = 'Shorten URL';
            submitBtn.removeAttribute('disabled')

            url.value = null;
        });
    });

    const shortenAnotherBtn = document.querySelector("#result-shorten-another");
    shortenAnotherBtn.addEventListener('click', () => {
        snipURLInput.value = "";
        longURLInput.value = "";
        snipURLAnchor.href = "#";

        form.classList.remove("d-none");
        resultTable.classList.add("d-none");

        document.querySelector(".btn-close").click();
    })

    const copySnipURLBtn = document.querySelector("#result-copy-snip-url");
    copySnipURLBtn.addEventListener('click', async () => {
        snipURLInput.select();
        snipURLInput.setSelectionRange(0, 99999); // For mobile devices

        await navigator.clipboard.writeText(snipURLInput.value);
    })
})()