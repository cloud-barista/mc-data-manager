/*!
    * Start Bootstrap - SB Admin v7.0.7 (https://startbootstrap.com/template/sb-admin)
    * Copyright 2013-2023 Start Bootstrap
    * Licensed under MIT (https://github.com/StartBootstrap/startbootstrap-sb-admin/blob/master/LICENSE)
    */
    // 
// Scripts
// 

window.addEventListener('DOMContentLoaded', event => {

    // Toggle the side navigation
    const sidebarToggle = document.body.querySelector('#sidebarToggle');
    if (sidebarToggle) {
        // Uncomment Below to persist sidebar toggle between refreshes
        // if (localStorage.getItem('sb|sidebar-toggle') === 'true') {
        //     document.body.classList.toggle('sb-sidenav-toggled');
        // }
        sidebarToggle.addEventListener('click', event => {
            event.preventDefault();
            document.body.classList.toggle('sb-sidenav-toggled');
            localStorage.setItem('sb|sidebar-toggle', document.body.classList.contains('sb-sidenav-toggled'));
        });
    }

    generateFormSubmit();
    
});


function generateFormSubmit() {
    const form = document.getElementById('genForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        loadingButtonOn();
        resultCollpase();

        const payload = new FormData(form);
        let jsonData = JSON.stringify(Object.fromEntries(payload));
        console.log(jsonData);
        const target= document.getElementById('genTarget').value;
        const url = "/generate/" + target;

        console.log(url);
   
        let req;
        if (target == "gcs" || target == "firestore") {
            req = { method: 'POST', body: payload };
        } else {
            req = { method: 'POST', body: jsonData };
        }

        fetch(url, req)
        .then(response => {
            return response.json();
        })
        .then(json => {
            const resultText = document.getElementById('resultText');
            resultText.value = json.Result;
            console.log(json);
            console.log("generate done.");
        })
        .catch(reason => {
            console.log(reason);
            alert(reason);
        })
        .finally(() => {
            loadingButtonOff();
        });

        console.log("generate progressing...");

    });
}

function loadingButtonOn() {
    let btn = document.getElementById('submitBtn');
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>&nbsp;진행 중..';
}

function loadingButtonOff() {
    let btn = document.getElementById('submitBtn');
    btn.disabled = false;
    btn.innerHTML = '생성';
}

function resultCollpase() {
    const colp = new bootstrap.Collapse('#resultCollapse', {
        toggle: true
    });
    colp.show();
}