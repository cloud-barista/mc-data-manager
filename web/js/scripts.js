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
        loadProfileList();
    }
    if (document.getElementById('migForm')) {
        migrationFormSubmit();
        loadProfileList();
        setPicker();
        setFilterAccordion();
    }
    if (document.getElementById('backForm')) {
        backUpFormSubmit();
        loadProfileList();
    }
    if (document.getElementById('restoreForm')) {
        RestoreFormSubmit();
        loadProfileList();
    }

    if (document.getElementById('credentialForm')) {
        setSelectBox();
        credentialFormSubmit();        
    }

});

function loadingButtonOn() {
    const resultText = document.getElementById('resultText');
    resultText.value = ""

    let btn = document.getElementById('submitBtn');
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>&nbsp;In progress..';
}

function loadingButtonOff() {
    let btn = document.getElementById('submitBtn');
    btn.disabled = false;
    btn.innerHTML = 'Submit';
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

        const requestBody = new Object();
        requestBody.targetPoint = {
            ...jsonData
        };
        console.log(JSON.stringify(jsonData))

        requestBody.targetPoint.provider = document.getElementById('targetPoint[provider]')?.value;
        const target = document.getElementById('genTarget').value;
        
        // 객체 형태 맞춤
        const raw = jsonData?.["targetPoint[credentialId]"]?? null
        requestBody.targetPoint.credentialId = raw != null ? parseInt(raw):raw;
        delete requestBody.targetPoint["targetPoint[credentialId]"]
        delete requestBody["targetPoint[credentialId]"]

        requestBody.targetPoint.region = jsonData?.["targetPoint[region]"]?? null
        delete requestBody.targetPoint["targetPoint[region]"]
        delete requestBody["targetPoint[region]"]
        
        if (requestBody.targetPoint.provider != "ncp") {
            delete requestBody.targetPoint.endpoint
        }

        requestBody.dummy = requestBody.targetPoint

        const url = "/generate/" + target;

        let req;

        req = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
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

function setSelectBox() {
    $("#select-credential-csp").change(function() {
        const selected = $(this).val();
        let formHtml = "";

        if (selected === "aws" || selected === "ncp") {
          formHtml = `
            <div class="input-group mb-3">
                <span class="input-group-text rounded-start"><i class="fa-solid fa-key"></i></span>
                <div class="form-floating">
                    <input type="text" class="form-control rounded-end" id="mig-aws-accessKey" name="accessKey" placeholder="Access Key" required>
                    <label for="mig-aws-accessKey">Access Key</label>
                </div>
            </div>

            <div class="input-group mb-3">
                <span class="input-group-text rounded-start"><i class="fa-solid fa-lock"></i></span>
                <div class="form-floating">
                    <input type="password" class="form-control rounded-end" id="mig-aws-secretKey" name="secretKey" placeholder="Secret Key" required>
                    <label for="mig-aws-secretKey">Secret Key</label>
                </div>
            </div>
          `;
        } else if (selected === "gcp") {
            formHtml = `
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-key"></i></span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="gcp-s3-accessKey" name="s3accessKey" placeholder="S3 Access Key" required>
                        <label for="gcp-s3-accessKey">S3 Access Key</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-lock"></i></span>
                    <div class="form-floating">
                        <input type="password" class="form-control rounded-end" id="gcp-s3-secretKey" name="s3secretKey" placeholder="S3 Secret Key" required>
                        <label for="gcp-s3-secretKey">S3 Secret Key</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start">Credential Json</span>
                    <div class="form-floating">                    
                        <textarea rows="10" class="form-control rounded-end" id="mig-gcp-json" name="gcpJson" placeholder="Input Credential Json" style="min-height: 300px; height: 300px" required></textarea>
                        <label for="mig-gcp-json">Input Credential Json</label>
                    </div>
                </div>
            `;
        }

        $("#credential-dynamicForm").html(formHtml);
      });
}

function setFilterAccordion() {
    $("#btnFilterExpand").click(function() {
        $("#filterContent").slideToggle("fast");
    });
}

function credentialFormSubmit() {
    const form = document.getElementById('credentialForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        loadingButtonOn();
        resultCollpase();

        const payload = new FormData(form);
        let tempObject = Object.fromEntries(payload);        
        let jsonData = {};        
        const { cspType, name, accessKey, secretKey, gcpJson } = tempObject;
        if (accessKey) {            
            jsonData = {
                cspType,
                name,
                credentialJson: {
                    accessKey,
                    secretKey
                }
            }
        } else {
            // const { cspType, name, type, project_id, private_key_id, private_key, client_email, client_id, auth_uri, token_uri, auth_provider_x509_cert_url, client_x509_cert_url, universe_domain } = tempObject;
            jsonData = {
                cspType,
                name,
                credentialJson: JSON.parse(gcpJson)
            }
        }
        // jsonData = convertCheckboxParams(jsonData)
        // jsonData.targetPoint = {
        //     ...jsonData
        // };
        console.log('credentialFormSubmit jsonData: ', jsonData)
        // jsonData.targetPoint.provider = document.getElementById('provider').value;
        // const target = document.getElementById('genTarget').value;
        // jsonData.dummy = jsonData.targetPoint
        // if ((jsonData.targetPoint.provider == "ncp") && (jsonData.targetPoint.endpoint == "")) {
        //     jsonData.targetPoint.endpoint = "https://kr.object.ncloudstorage.com"
        // }
        const url = "/credentials";

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
                console.log(response);
                
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(json => {
                const resultText = document.getElementById('resultText');
                // resultText.value = json.Result;
                resultText.value = json && 'Success';
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

function setPicker() {
    $( function() {
        // $("#datepicker1").datepicker();
        // $("#datepicker2").datepicker();
        $("#datepicker1").datetimepicker({ 
            format: "Y-m-d H:i:s",
            step: 1,
        });
        $("#datepicker2").datetimepicker({ 
            format: "Y-m-d H:i:s",
            step: 1,
        });
    } );
}

function loadProfileList() {
    let url = "/credentials";

    let req;

        req = {
            method: 'GET',
            // headers: {
            //     'Content-Type': 'application/json'
            // },
            // body: JSON.stringify(jsonData)
        };

        fetch(url, req)
            .then(response => {
                return response.json();
            })
            .then(json => {
                // const resultText = document.getElementById('resultText');
                // resultText.value = json.Result;
                // console.log(json);

                const options = json
                .map((item) => {
                    return {
                        label: `${item.cspType.toUpperCase()} - ${item.name}`,
                        value: item.credentialId,
                        provider: item.cspType
                    }
                });

                const sourceSelectElement = document.getElementById("sourceCredentialSelect");

                if (sourceSelectElement) {
                    // placeholder 역할 옵션 추가
                    // const placeholder = document.createElement("option");
                    // placeholder.textContent = "Select Credential";
                    // placeholder.disabled = true;
                    // placeholder.selected = true;
                    // awsSelect.appendChild(placeholder);

                    options.forEach(optionData => {
                        const option = document.createElement("option");
                        option.value = optionData.value;
                        option.textContent = optionData.label;
                        option.setAttribute("data-provider", optionData.provider);
                        sourceSelectElement.appendChild(option);
                    });
                }

                const targetSelectElement = document.getElementById("targetCredentialSelect");

                if (targetSelectElement) {
                    // placeholder 역할 옵션 추가
                    // const placeholder = document.createElement("option");
                    // placeholder.textContent = "Select Credential";
                    // placeholder.disabled = true;
                    // placeholder.selected = true;
                    // awsSelect.appendChild(placeholder);

                    options.forEach(optionData => {
                        const option = document.createElement("option");
                        option.value = optionData.value;
                        option.textContent = optionData.label;
                        option.setAttribute("data-provider", optionData.provider);
                        targetSelectElement.appendChild(option);
                    });
                }

                // console.log('options: ', options);

                const capSelect = document.getElementById("mig-filter-sizeFilteringUnit");

                const capOptions = [
                    {
                        label: 'KB',
                        value: 'KB',
                    },
                    {
                        label: 'MB',
                        value: 'MB',
                    },
                    {
                        label: 'GB',
                        value: 'GB',
                    },
                ]

                if (capSelect) {
                    // placeholder 역할 옵션 추가
                    // const placeholder = document.createElement("option");
                    // placeholder.textContent = "Select Unit";
                    // placeholder.disabled = true;
                    // placeholder.selected = true;
                    // capSelect.appendChild(placeholder);

                    capOptions.forEach(optionData => {
                        const option = document.createElement("option");
                        option.value = optionData.value;
                        option.textContent = optionData.label;
                        capSelect.appendChild(option);
                    });
                }
                

                // window.mcmpData = {                    
                //     profileList: json,
                //     profileOptions: options
                // };                

                // console.log('window mcmpData: ', window.mcmpData);
                
                // console.log("migration done.");
            })
            .catch(reason => {
                console.log(reason);
                alert(reason);
            });
}

function migrationFormSubmit() {
    const form = document.getElementById('migForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const payload = new FormData(form);
        let jsonData = formDataToObject(payload)
        if(jsonData.sourcePoint.credentialId == "none") {
            alert("source credential not selected");
            return
        }
        if(jsonData.targetPoint.credentialId == "none") {
            alert("target credential not selected");
            return
        }
        if(jsonData.targetPoint.bucket == "none" || jsonData.targetPoint.bucket == "-" || jsonData.targetPoint.bucket == ""
            || jsonData.sourcePoint.bucket == "none" || jsonData.sourcePoint.bucket == "-" || jsonData.sourcePoint.bucket == ""
        ) {
            alert("please select or create bucket");
            return
        }

        loadingButtonOn();
        resultCollpase();
        console.log(jsonData)
        const service = document.getElementById('migService').value;
        jsonData.targetPoint.provider = getInputValue('targetPoint[provider]');
        jsonData.sourcePoint.provider = getInputValue('sourcePoint[provider]');
        jsonData.targetPoint.credentialId = parseInt(jsonData.targetPoint.credentialId);
        jsonData.sourcePoint.credentialId = parseInt(jsonData.sourcePoint.credentialId);

        if(service == "objectstorage") {
            applyFilter(jsonData)
        }

        let url = "/migrate/" + service;

        if ((jsonData.targetPoint.provider == "ncp") || (jsonData.targetPoint.endpoint == "")) {
            jsonData.targetPoint.endpoint = "https://kr.object.ncloudstorage.com"
        }
        if ((jsonData.sourcePoint.provider == "ncp") || (jsonData.sourcePoint.endpoint == "")) {
            jsonData.sourcePoint.endpoint = "https://kr.object.ncloudstorage.com"
        }

        console.log('jsonData: ', jsonData);

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

function applyFilter(jsonData) {
    if (jsonData.sourceFilter.path === "" || jsonData.sourceFilter.path === null) {
        jsonData.sourceFilter.path = null;
    } 

    if (jsonData.sourceFilter.minSize === "" || jsonData.sourceFilter.minSize === null) {
        jsonData.sourceFilter.minSize = null;
    } else {
        jsonData.sourceFilter.minSize = parseFloat(jsonData.sourceFilter.minSize);
    }

    if (jsonData.sourceFilter.maxSize === "" || jsonData.sourceFilter.maxSize === null) {
        jsonData.sourceFilter.maxSize = null;
    } else {
        jsonData.sourceFilter.maxSize = parseFloat(jsonData.sourceFilter.maxSize);
    }
    
    if (jsonData.sourceFilter.modifiedAfter && jsonData.sourceFilter.modifiedAfter.trim() !== "") {
        jsonData.sourceFilter.modifiedAfter = jsonData.sourceFilter.modifiedAfter.replace(" ", "T") + "+09:00";
    } else {
        jsonData.sourceFilter.modifiedAfter = null;
    }
    
    if (jsonData.sourceFilter.modifiedBefore && jsonData.sourceFilter.modifiedBefore.trim() !== "") {
        jsonData.sourceFilter.modifiedBefore = jsonData.sourceFilter.modifiedBefore.replace(" ", "T") + "+09:00";
    } else {
        jsonData.sourceFilter.modifiedBefore = null;
    }

    if (jsonData.sourceFilter.contains === "") {
        jsonData.sourceFilter.contains = null;
    } else {
        jsonData.sourceFilter.contains = jsonData.sourceFilter.contains.replace(/ /g,"").split(',');
    }

    if (jsonData.sourceFilter.suffixes === "") {
        jsonData.sourceFilter.suffixes = null;
    }
    if (jsonData.sourceFilter.exact === "") {
        jsonData.sourceFilter.exact = null;
    }

    console.log(
        "minSize type:", typeof jsonData.sourceFilter.minSize, 
        "value:", jsonData.sourceFilter.minSize
    );
    console.log(
        "maxSize type:", typeof jsonData.sourceFilter.maxSize, 
        "value:", jsonData.sourceFilter.maxSize
    );

    console.log(
        "contains type:", typeof jsonData.sourceFilter.contains, 
        "value:", jsonData.sourceFilter.contains
    );
}

function backUpFormSubmit() {
    const form = document.getElementById('backForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const payload = new FormData(form);
        let jsonData = formDataToObject(payload)

        const provider = document.getElementById('sourcePoint[provider]').value;
        const service = document.getElementById('srcService').value;
        if(service != "rdbms" && jsonData.sourcePoint.credentialId == "none") {
            alert("credential not selected");
            return
        }
        // console.log(provider)

        loadingButtonOn();
        resultCollpase();

        let url = "/backup/" + service;
        console.log(url);

        jsonData.sourcePoint.credentialId = parseInt(jsonData.sourcePoint.credentialId);
        jsonData.sourcePoint.provider = provider
        // console.log(jsonData)

        if (jsonData.sourcePoint.provider != "ncp") {
            delete jsonData.sourcePoint.endpoint
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

        const payload = new FormData(form);
        let jsonData = formDataToObject(payload)

        const provider = document.getElementById('targetPoint[provider]').value;
        const service = document.getElementById('srcService').value;
        if(service != "rdbms" && jsonData.targetPoint.credentialId == "none") {
            alert("credential not selected");
            return
        }

        loadingButtonOn();
        resultCollpase();

        let url = "/restore/" + service;
        console.log(url);
        console.log(jsonData)

        jsonData.targetPoint.credentialId = parseInt(jsonData.targetPoint.credentialId);
        jsonData.targetPoint.provider = provider

        if (jsonData.targetPoint.provider != "ncp") {
            delete jsonData.targetPoint.endpoint
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

            const userConfirmed = confirm('Are you sure you want to clear all services?');
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
                        alert('The service has been successfully cleared.');
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || 'An error occurred while clearing the service.');
                        });
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert(`Error: ${error.message}`);
                });
        });
    } 
});





document.addEventListener('DOMContentLoaded', function () {
    const genServiceLink = document.getElementById('genServiceLink');

    if (genServiceLink) {
        genServiceLink.addEventListener('click', function (event) {
            event.preventDefault();

            const userConfirmed = confirm('Are you sure you want to create a data-related service?');
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
                        alert('Service creation request has been submitted.');
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || 'An error occurred while submitting the service creation request.');
                        });
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert(`Error: ${error.message}`);
                });
        });
    } else {
        console.error('Gen Service link not found.');
    }
});


document.addEventListener('DOMContentLoaded', function () {
    const delServiceLink = document.getElementById('delServiceLink');

    if (delServiceLink) {
        delServiceLink.addEventListener('click', function (event) {
            event.preventDefault();

            const userConfirmed = confirm('Are you sure you want to remove the data-related service?');
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
                        alert('Service removal request has been submitted.');
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || 'An error occurred while submitting the service removal request.');
                        });
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert(`Error: ${error.message}`);
                });
        });
    } else {
        console.error('Del Service link not found.');
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



// @@  
function updateRegionsByProvider(prefix, provider) {
    const select = document.getElementById(prefix + "RegionSelect");
    if (!select) return;
  
    // 모든 optgroup 숨기기
    select.querySelectorAll("optgroup").forEach(group => {
      group.style.display = "none";
    });

    if(provider == "none") {
      select.value = "none"
      return
    }
  
    // 해당 provider optgroup만 보이기
    const target = select.querySelector(`optgroup[data-provider="${provider}"]`);
    if (target) {
      target.style.display = "";
      // 첫 option 선택
      select.value = target.querySelector("option")?.value || "";
    }
}


function updateLabelByProvider(prefix, provider) {
    const service = document.getElementById("migService");
    const select = document.getElementById(prefix + "PointLabel");
    if (!service || !select) return;
  
    // 모든 optgroup 숨기기
    if(provider == "none") {
        select.value = "-"
        return
    }
    
    // label 갱신
    select.value = getServiceName(service.value, provider);
    
    // region/mongo 영역 토글
    toggleDivs(prefix, provider, service.value);
}

function toggleDivs(prefix, provider, service) {
    const regionDiv = document.getElementById(prefix + "RegionDiv");
    const mongoDiv = document.getElementById(prefix + "MongoDiv");
    if (!regionDiv || !mongoDiv) return;

    const showMongo = service === "nrdbms" && provider === "ncp";

    regionDiv.style.display = showMongo ? "none" : "";
    mongoDiv.style.display = showMongo ? "" : "none";

    const regionSelect = regionDiv.querySelector("select");
    if (regionSelect) {
        regionSelect.disabled = showMongo; // mongoDiv 보이면 regionSelect 비활성화
    }
      // mongoDiv 하위 모든 input 제어
    const mongoInputs = mongoDiv.querySelectorAll("input");
    mongoInputs.forEach(input => {
        input.disabled = !showMongo; // mongoDiv 보일 때만 활성화
    });
}

document.addEventListener("DOMContentLoaded", () => {
    const initProviderHandlers = (prefix) => {
        const credSelect = document.getElementById(prefix + "CredentialSelect");
        const providerInput = document.getElementById(prefix + "Point[provider]");
        if (!credSelect || !providerInput) return;

        // 초기 렌더링 시 region 갱신
        const initialProvider = credSelect.selectedOptions[0]?.dataset?.provider || "none";
        providerInput.value = initialProvider;
        updateRegionsByProvider(prefix, initialProvider);
        updateLabelByProvider(prefix, initialProvider);

        // 이벤트 핸들러 등록
        credSelect.addEventListener("change", (e) => {
            const provider = e.target.selectedOptions[0]?.dataset?.provider || "none";
            providerInput.value = provider;
            providerInput.dispatchEvent(new Event('change'));
            updateRegionsByProvider(prefix, provider);
            updateLabelByProvider(prefix, provider);
        });
    };

    // source/target 공통 처리
    ["source", "target"].forEach(initProviderHandlers);
});

function ucFirst(str) {
    if (!str) {
      return "";
    }
    return str.charAt(0).toUpperCase() + str.slice(1);
}

function getServiceName(service, provider) {
    if (provider === "none") return "-";
  
    const serviceMap = {
      nrdbms: {
        aws: "AWS DynamoDB",
        ncp: "Naver MongoDB",
        gcp: "Google Firestore",
      },
      objectstorage: {
        aws: "AWS S3",
        ncp: "Naver Object Storage",
        gcp: "Google Cloud Storage",
      },
      rdbms: {
        default: "MySQL",
      },
    };
  
    const providers = serviceMap[service];
    if (!providers) return "-";
  
    // 우선순위: provider별 매핑 → 없으면 default → 없으면 "-"
    return providers[provider] || providers.default || "-";
}
