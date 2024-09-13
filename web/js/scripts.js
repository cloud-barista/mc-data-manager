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

    if (document.getElementById('genForm')) {
        generateFormSubmit();
    }
    if (document.getElementById('migForm')) {
        migrationFormSubmit();
    }
    if (document.getElementById('backForm')) {
        backUpFormSubmit();
    }
    
});

function convertCheckboxParams(obj) {
    for (const key in obj) {
        if (obj.hasOwnProperty(key)) {
            if (obj[key] === "on") {
                obj[key] = true;
            } else if (obj[key] === "off") {
                obj[key] = false;
            } else if (typeof obj[key] === "object" && !Array.isArray(obj[key])) {
                convertCheckboxParams(obj[key]);
            }
        }
    }
    return obj;
}

function formDataToObject(formData) {
    const data = {};
    formData.forEach((value, key) => {
      const match = key.match(/(\w+)\[(\w+)\]/);
      if (match) {
        const objName = match[1];
        const paramName = match[2];
        if (!data[objName]) {
          data[objName] = {};
        }
        data[objName][paramName] = value;
      } else {
        data[key] = value;
      }
    });
    return data;
  }

  function getInputValue(id) {
    const element = document.getElementById(id);
    if (!element) {
    //   console.warn(`Element with id '${id}' not found.`);
      return null;
    }
  
    const value = element.value.trim();
    return value !== "" ? value : null;
  }


function generateFormSubmit() {
    const form = document.getElementById('genForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        loadingButtonOn();
        resultCollpase();

        const payload = new FormData(form);
        let jsonData = Object.fromEntries(payload);
        jsonData=convertCheckboxParams(jsonData)
        jsonData.targetPoint = {
            ...jsonData
        };
        console.log(jsonData)
        jsonData.targetPoint.provider = document.getElementById('provider').value;
        const target = document.getElementById('genTarget').value;

        if ( (jsonData.targetPoint.provider =="ncp") && (jsonData.targetPoint.endpoint =="") ) {
            jsonData.targetPoint.endpoint ="https://kr.object.ncloudstorage.com"
        }
        const url = "/generate/" + target;
        
        let req;

        req = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(jsonData)
        };

        fetch(url, req)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(json => {
            const resultText = document.getElementById('resultText');
            resultText.value = json.Result;
            console.log(json);
            console.log("Generate done.");
        })
        .catch(reason => {
            console.error("Error during generate:", reason);
            alert(reason.message || reason);
        })
        .finally(() => {
            loadingButtonOff();
        });

        console.log("Generate progressing...");
    });
}

function migrationFormSubmit() {
    const form = document.getElementById('migForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        loadingButtonOn();
        resultCollpase();

        const payload = new FormData(form);
        let jsonData= formDataToObject(payload)
        console.log(jsonData)
        const dest = document.getElementById('migDest').value;
        const source = document.getElementById('migSource').value;
        jsonData.targetPoint.provider = getInputValue('targetPoint[provider]');
        jsonData.sourcePoint.provider = getInputValue('sourcePoint[provider]');


        let url = "/migration/" + source;
        if (source != dest) {
            url = url + "/" + dest;
        }

        if ( (jsonData.targetPoint.provider =="ncp") && (jsonData.targetPoint.endpoint =="") ) {
            jsonData.targetPoint.endpoint ="https://kr.object.ncloudstorage.com"
        }
        if ( (jsonData.sourcePoint.provider =="ncp") && (jsonData.sourcePoint.endpoint =="") ) {
            jsonData.sourcePoint.endpoint ="https://kr.object.ncloudstorage.com"
        }

        console.log(url);

        let req;

        req = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(jsonData)
        };

        fetch(url, req)
        .then(response => {
            return response.json();
        })
        .then(json => {
            const resultText = document.getElementById('resultText');
            resultText.value = json.Result;
            console.log(json);
            console.log("migration done.");
        })
        .catch(reason => {
            console.log(reason);
            alert(reason);
        })
        .finally(() => {
            loadingButtonOff();
        });

        console.log("migration progressing...");
    });
}

function backUpFormSubmit() {
    const form = document.getElementById('backForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        loadingButtonOn();
        resultCollpase();

        const payload = new FormData(form);
        const dest = document.getElementById('backDest').value;
        const source = document.getElementById('backSource').value;
        let url = "/backup/" + source;

        console.log(url);

        fetch(url, {
            method: 'POST',
            body: payload
        })
        .then(response => {
            return response.json();
        })
        .then(json => {
            const resultText = document.getElementById('resultText');
            resultText.value = json.Result;
            console.log(json);
            console.log("backup done.");
        })
        .catch(reason => {
            console.log(reason);
            alert(reason);
        })
        .finally(() => {
            loadingButtonOff();
        });

        console.log("backup progressing...");
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


function test() {

}