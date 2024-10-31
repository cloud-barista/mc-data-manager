/*!
    * Start Bootstrap - SB Admin v7.0.7 (https://startbootstrap.com/template/sb-admin)
    * Copyright 2013-2023 Start Bootstrap
    * Licensed under MIT (https://github.com/StartBootstrap/startbootstrap-sb-admin/blob/master/LICENSE)
    */
// 
// Scripts
// 




// for mc-data-manger

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
    if (document.getElementById('restoreForm')) {
        RestoreFormSubmit();
    }

});

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
        jsonData = convertCheckboxParams(jsonData)
        jsonData.targetPoint = {
            ...jsonData
        };
        console.log(jsonData)
        jsonData.targetPoint.provider = document.getElementById('provider').value;
        const target = document.getElementById('genTarget').value;
        jsonData.dummy = jsonData.targetPoint
        if ((jsonData.targetPoint.provider == "ncp") && (jsonData.targetPoint.endpoint == "")) {
            jsonData.targetPoint.endpoint = "https://kr.object.ncloudstorage.com"
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
        let jsonData = formDataToObject(payload)
        console.log(jsonData)
        const dest = document.getElementById('migDest').value;
        const source = document.getElementById('migSource').value;
        const service = document.getElementById('migService').value;
        jsonData.targetPoint.provider = getInputValue('targetPoint[provider]');
        jsonData.sourcePoint.provider = getInputValue('sourcePoint[provider]');


        let url = "/migrate/" + service;

        if ((jsonData.targetPoint.provider == "ncp") && (jsonData.targetPoint.endpoint == "")) {
            jsonData.targetPoint.endpoint = "https://kr.object.ncloudstorage.com"
        }
        if ((jsonData.sourcePoint.provider == "ncp") && (jsonData.sourcePoint.endpoint == "")) {
            jsonData.sourcePoint.endpoint = "https://kr.object.ncloudstorage.com"
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

        var service = document.getElementById('srcService').value;
        let url = "/backup/" + service;
        console.log(url);


        let jsonData = formDataToObject(payload)
        console.log(jsonData)


        if ((jsonData.targetPoint.provider == "ncp") && (jsonData.targetPoint.endpoint == "")) {
            jsonData.targetPoint.endpoint = "https://kr.object.ncloudstorage.com"
        }
        if ((jsonData.sourcePoint.provider == "ncp") && (jsonData.sourcePoint.endpoint == "")) {
            jsonData.sourcePoint.endpoint = "https://kr.object.ncloudstorage.com"
        }

        let req = {
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

function RestoreFormSubmit() {
    const form = document.getElementById('restoreForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        loadingButtonOn();
        resultCollpase();

        const payload = new FormData(form);

        var service = document.getElementById('srcService').value;
        let url = "/restore/" + service;
        console.log(url);


        let jsonData = formDataToObject(payload)
        console.log(jsonData)


        if ((jsonData.targetPoint.provider == "ncp") && (jsonData.targetPoint.endpoint == "")) {
            jsonData.targetPoint.endpoint = "https://kr.object.ncloudstorage.com"
        }
        if ((jsonData.sourcePoint.provider == "ncp") && (jsonData.sourcePoint.endpoint == "")) {
            jsonData.sourcePoint.endpoint = "https://kr.object.ncloudstorage.com"
        }

        let req = {
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
                console.log("restore done.");
            })
            .catch(reason => {
                console.log(reason);
                alert(reason);
            })
            .finally(() => {
                loadingButtonOff();
            });

        console.log("restore progressing...");
    });
}

document.addEventListener('DOMContentLoaded', function () {
    const clearServiceLink = document.getElementById('clearServiceLink');

    if (clearServiceLink) {
        clearServiceLink.addEventListener('click', function (event) {
            event.preventDefault();

            const userConfirmed = confirm('정말로 서비스를 클리어하시겠습니까?');
            if (!userConfirmed) {
                return;
            }

            fetch('/service/clearAll', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
                .then(response => {
                    if (response.ok) {
                        alert('서비스가 성공적으로 클리어되었습니다.');
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || '서비스 클리어 중 오류가 발생했습니다.');
                        });
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert(`오류: ${error.message}`);
                });
        });
    } else {
        console.error('Clear Service 링크를 찾을 수 없습니다.');
    }
});





document.addEventListener('DOMContentLoaded', function () {
    const genServiceLink = document.getElementById('genServiceLink');

    if (genServiceLink) {
        genServiceLink.addEventListener('click', function (event) {
            event.preventDefault();

            const userConfirmed = confirm('정말로 데이터 관련 서비스를 생성하시겠습니까?');
            if (!userConfirmed) {
                return;
            }

            fetch('/service/apply', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
                .then(response => {
                    if (response.ok) {
                        alert('서비스 생성 요청이 전달 되었습니다.');
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || '서비스 생성 요청 중 오류가 발생했습니다.');
                        });
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert(`오류: ${error.message}`);
                });
        });
    } else {
        console.error('Gen Service 링크를 찾을 수 없습니다.');
    }
});


document.addEventListener('DOMContentLoaded', function () {
    const delServiceLink = document.getElementById('delServiceLink');

    if (delServiceLink) {
        delServiceLink.addEventListener('click', function (event) {
            event.preventDefault();

            const userConfirmed = confirm('정말로 데이터 관련 서비스를 제거하시겠습니까?');
            if (!userConfirmed) {
                return;
            }

            fetch('/service/destroy', {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
                .then(response => {
                    if (response.ok) {
                        alert('서비스 제거 요청이 전달 되었습니다.');
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || '서비스 제거 요청 중 오류가 발생했습니다.');
                        });
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert(`오류: ${error.message}`);
                });
        });
    } else {
        console.error('Del Service 링크를 찾을 수 없습니다.');
    }
});

// for m-cmp/mc-web-console

// message  Info
// {
//     accessToken: "accesstokenExample",
//     workspaceInfo: {
//         "id": "8b2df1f9-b937-4861-b5ce-855a41c346bc",
//         "name": "workspace2",
//         "description": "workspace2 desc",
//         "created_at": "2024-06-18T00:10:16.192337Z",
//         "updated_at": "2024-06-18T00:10:16.192337Z"
//     },
//     projectInfo: {
//         "id": "1e88f4ea-d052-4314-80a4-9ac3f6691feb",
//         "ns_id": "project1",
//         "name": "project1",
//         "description": "project1 desc",
//         "created_at": "2024-06-18T00:28:57.094105Z",
//         "updated_at": "2024-06-18T00:28:57.094105Z"
//     },
//     operationId: "abc"
// };

window.addEventListener("message", async function (event) {
    const data = event.data;
    console.log("iframeServer : Message received :", data);
    try {
        // const nsId = data.projectInfo.ns_id
        // business logic 

    } catch (error) {
        console.error("Error in processing message:", error);
    }
});