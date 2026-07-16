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
        setPicker();
        setFilterAccordion();
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

// migrate/backup/restore 제출 시 opId를 sessionStorage에 저장하고, 페이지 복귀 시 1회만
// GET /{kind}/{id}로 상태를 조회해 복원한다 (반복 폴링 없음).
const TaskProgress = (() => {
    const KEY = 'mcdm.taskProgress';        // { kind, opId }

    function read() { try { return JSON.parse(sessionStorage.getItem(KEY) || 'null'); } catch (e) { return null; } }
    function save(kind, opId) { try { sessionStorage.setItem(KEY, JSON.stringify({ kind, opId })); } catch (e) {} }
    function clear() { try { sessionStorage.removeItem(KEY); } catch (e) {} }
    function newOpId(kind) { return kind + '-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8); }

    // 완료/실패 → onDone + 저장 삭제 / 진행 중 → onRunning / 없음·오류 → 저장 삭제
    function resume(expectedKind, onRunning, onDone) {
        const s = read();
        if (!s || s.kind !== expectedKind || !s.opId) return;
        fetch('/' + expectedKind + '/' + encodeURIComponent(s.opId))
            .then(r => r.ok ? r.json() : Promise.reject(r.status))
            .then(task => {
                const st = (task.status || '').toString().toLowerCase();
                if (st === 'completed') { clear(); onDone('Completed.'); }
                else if (st === 'failed') { clear(); onDone('Failed. Check the server logs.'); }
                else { onRunning(); }
            })
            .catch(() => clear());
    }

    function setResult(text) {
        resultCollpase();
        const rt = document.getElementById('resultText');
        if (rt) rt.value = text;
    }

    // migrate/backup/restore가 쓰는 표준 복원 UI — 셋 다 동작이 같아 여기에 둔다.
    // 진행 중이면 스피너 + 안내, 끝났으면 결과 문구를 Result 패널에 넣는다.
    function attach(kind) {
        resume(kind,
            () => { loadingButtonOn(); setResult('In progress.. (refresh to update the status)'); },
            text => { setResult(text); loadingButtonOff(); });
    }

    // 표준 제출 — opId를 심어 저장한 뒤 POST한다.
    // 작업이 수십 분 걸릴 수 있어 요청이 커넥션 타임아웃으로 끊길 수 있다. 그 경우
    // catch에서 저장을 지우지 않고 남겨, 새로고침 시 attach(kind)가 상태를 복원한다.
    function submit(kind, url, jsonData) {
        const opId = newOpId(kind);
        jsonData.operationId = opId;
        save(kind, opId);

        return fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(jsonData),
        })
            .then(response => response.json().catch(() => ({})))
            .then(json => { setResult((json && json.Result) || ''); clear(); })
            .catch(reason => { console.error(kind + ' request dropped:', reason); })
            .finally(() => { loadingButtonOff(); });
    }

    return { newOpId, save, clear, resume, attach, submit };
})();

// Result 패널을 항상 "펼침"으로 (toggle이 아니라 show — 이미 펼쳐져 있으면 유지).
// 기존 toggle:true는 재호출 시 오히려 접히거나 인스턴스 중복으로 불안정했다.
function resultCollpase() {
    const el = document.getElementById('resultCollapse');
    if (!el) return;
    bootstrap.Collapse.getOrCreateInstance(el, { toggle: false }).show();
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

const NCP_OBJECTSTORAGE_ENDPOINT = "https://kr.object.ncloudstorage.com";

function resolveAlibabaEndpoint(region) {
    if (!region) {
        return "";
    }
    const normalized = region.trim().toLowerCase();
    return `https://oss-${normalized}.aliyuncs.com`;
}

// ncp/alibaba는 endpoint가 필수인데 폼에 비어 있을 수 있어 provider 규칙으로 채운다.
// 그 외 provider는 endpoint 개념이 없다 → 호출부에서 제거한다.
// point: { provider, region, endpoint } 를 제자리에서 수정한다.
function ensureEndpoint(point) {
    if (!point || point.endpoint) return;
    if (point.provider === "ncp") {
        point.endpoint = NCP_OBJECTSTORAGE_ENDPOINT;
    } else if (point.provider === "alibaba") {
        point.endpoint = resolveAlibabaEndpoint(point.region);
    }
}

// 버킷 리전 후보들을 Region 셀렉트 옵션과 대소문자 무시로 매칭해 실제 옵션 값을 돌려준다.
// 없으면 null. 후보를 앞에서부터 순서대로 시도한다 (예: assignedRegion → connectionName 파생).
// NCP는 assignedRegion이 옵션과 형식이 달라(KR vs kr-standard) connectionName 파생값이 매칭된다.
function matchRegionOption(selectEl, ...candidates) {
    if (!selectEl) return null;
    for (const c of candidates) {
        const t = (c || '').trim().toLowerCase();
        if (!t) continue;
        const opt = [...selectEl.options].find(o => (o.value || '').trim().toLowerCase() === t);
        if (opt) return opt.value;
    }
    return null;
}

function generateFormSubmit() {
    const form = document.getElementById('genForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const target = document.getElementById('genTarget').value;
        const payload = new FormData(form);
        let jsonData = Object.fromEntries(payload);
        jsonData = convertCheckboxParams(jsonData)

        const bucket = document.getElementById('newBucket')?.value;
        if (
            target === "objectstorage" &&
            (!bucket || ["none", "-", ""].includes(bucket.trim()))
        ) {
            alert("please select or create bucket");
            return
        }

        loadingButtonOn();
        resultCollpase();

        const requestBody = new Object();
        requestBody.targetPoint = {
            ...jsonData
        };

        requestBody.targetPoint.provider = document.getElementById('targetPoint[provider]')?.value;
        
        // 객체 형태 맞춤
        const raw = jsonData?.["targetPoint[credentialId]"]?? null
        requestBody.targetPoint.credentialId = raw != null ? parseInt(raw):raw;
        delete requestBody.targetPoint["targetPoint[credentialId]"]
        delete requestBody["targetPoint[credentialId]"]

        requestBody.targetPoint.region = jsonData?.["targetPoint[region]"]?? null
        delete requestBody.targetPoint["targetPoint[region]"]
        delete requestBody["targetPoint[region]"]

        requestBody.targetPoint.bucket = jsonData?.["targetPoint[bucket]"]?? null
        delete requestBody.targetPoint["targetPoint[bucket]"]
        delete requestBody["targetPoint[bucket]"]
        const endpointRaw = jsonData?.["targetPoint[endpoint]"];
        if (endpointRaw !== undefined) {
            requestBody.targetPoint.endpoint = endpointRaw || null;
            delete requestBody.targetPoint["targetPoint[endpoint]"];
            delete requestBody["targetPoint[endpoint]"];
        }
        
        ensureEndpoint(requestBody.targetPoint);
        if (!["ncp", "alibaba"].includes(requestBody.targetPoint.provider)) {
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
                document.getElementById('resultText').value = json.Result;
            })
            .catch(reason => {
                console.error("Error during generate:", reason);
                alert(reason.message || reason);
            })
            .finally(() => {
                loadingButtonOff();
            });
    });
}

function setSelectBox() {
    $("#select-credential-csp").change(function() {
        const selected = $(this).val();
        let formHtml = "";

        if (selected === "aws" || selected === "ncp" || selected === "alibaba") {
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
        } else if (selected === "ibm") {
            formHtml = `
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-key"></i></span>
                    <div class="form-floating">
                        <input type="password" class="form-control rounded-end" id="ibm-apiKey" name="apiKey" placeholder="API Key" required>
                        <label for="ibm-apiKey">API Key</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-key"></i></span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="ibm-s3AccessKey" name="s3AccessKey" placeholder="S3 Access Key">
                        <label for="ibm-s3AccessKey">S3 Access Key</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-lock"></i></span>
                    <div class="form-floating">
                        <input type="password" class="form-control rounded-end" id="ibm-s3SecretKey" name="s3SecretKey" placeholder="S3 Secret Key">
                        <label for="ibm-s3SecretKey">S3 Secret Key</label>
                    </div>
                </div>
            `;
        } else if (selected === "kt") {
            formHtml = `
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-user"></i></span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="kt-username" name="username" placeholder="Username" required>
                        <label for="kt-username">Username</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-lock"></i></span>
                    <div class="form-floating">
                        <input type="password" class="form-control rounded-end" id="kt-password" name="password" placeholder="Password" required>
                        <label for="kt-password">Password</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start">Domain Name</span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="kt-domainName" name="domainName" placeholder="Domain Name" required>
                        <label for="kt-domainName">Domain Name</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start">Project ID</span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="kt-projectID" name="projectID" placeholder="Project ID" required>
                        <label for="kt-projectID">Project ID</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-key"></i></span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="kt-s3AccessKey" name="s3AccessKey" placeholder="S3 Access Key">
                        <label for="kt-s3AccessKey">S3 Access Key</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-lock"></i></span>
                    <div class="form-floating">
                        <input type="password" class="form-control rounded-end" id="kt-s3SecretKey" name="s3SecretKey" placeholder="S3 Secret Key">
                        <label for="kt-s3SecretKey">S3 Secret Key</label>
                    </div>
                </div>
            `;
        } else if (selected === "tencent") {
            formHtml = `
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-id-card"></i></span>
                    <div class="form-floating">
                        <input type="text" class="form-control rounded-end" id="tencent-secretId" name="secretId" placeholder="Secret ID" required>
                        <label for="tencent-secretId">Secret ID</label>
                    </div>
                </div>
                <div class="input-group mb-3">
                    <span class="input-group-text rounded-start"><i class="fa-solid fa-lock"></i></span>
                    <div class="form-floating">
                        <input type="password" class="form-control rounded-end" id="tencent-secretKey" name="secretKey" placeholder="Secret Key" required>
                        <label for="tencent-secretKey">Secret Key</label>
                    </div>
                </div>
            `;
        }

        $("#credential-dynamicForm").html(formHtml);
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
        const { cspType, name } = tempObject;

        if (cspType === "aws" || cspType === "ncp" || cspType === "alibaba") {
            jsonData = {
                cspType,
                name,
                credentialJson: {
                    accessKey: tempObject.accessKey,
                    secretKey: tempObject.secretKey
                }
            };
        } else if (cspType === "gcp") {
            jsonData = {
                cspType,
                name,
                credentialJson: JSON.parse(tempObject.gcpJson)
            };
        } else if (cspType === "ibm") {
            jsonData = {
                cspType,
                name,
                credentialJson: {
                    apiKey: tempObject.apiKey,
                    s3AccessKey: tempObject.s3AccessKey || "",
                    s3SecretKey: tempObject.s3SecretKey || ""
                }
            };
        } else if (cspType === "kt") {
            jsonData = {
                cspType,
                name,
                credentialJson: {
                    identityEndpoint: tempObject.identityEndpoint,
                    username: tempObject.username,
                    password: tempObject.password,
                    domainName: tempObject.domainName,
                    projectID: tempObject.projectID,
                    s3AccessKey: tempObject.s3AccessKey || "",
                    s3SecretKey: tempObject.s3SecretKey || ""
                }
            };
        } else if (cspType === "tencent") {
            jsonData = {
                cspType,
                name,
                credentialJson: {
                    secretId: tempObject.secretId,
                    secretKey: tempObject.secretKey
                }
            };
        }
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
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(json => {
                document.getElementById('resultText').value = json && 'Success';
            })
            .catch(reason => {
                console.error("Error during credential submit:", reason);
                alert(reason.message || reason);
            })
            .finally(() => {
                loadingButtonOff();
            });
    });
}

function setFilterAccordion() {
    $("#btnFilterExpand").click(function() {
        $("#filterContent").slideToggle("fast");
    });
}

function setPicker() {
    $( function() {
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

// RDB·NoSQL 인스턴스 관리를 지원하지 않는 provider (SQL Database / No-SQL 화면에서 제외)
const DB_EXCLUDED_PROVIDERS = ['ibm', 'nhn', 'kt', 'tencent', 'azure', 'openstack'];
// Object Storage 계열 화면(비-DB)에서 미지원 provider 제외
const OBJECTSTORAGE_EXCLUDED_PROVIDERS = ['azure', 'openstack'];

// backup/restore처럼 한 페이지에서 서비스를 전환하는 화면용:
// SQL/No-SQL 서비스 선택 시 미지원 provider 옵션을 숨기고, 이미 선택돼 있었으면 선택을 해제한다.
// Object Storage로 돌아오면(isDbService=false) 전체 provider를 다시 노출한다
function applyDbProviderExclusion(selectEl, isDbService) {
    if (!selectEl) return;
    [...selectEl.options].forEach(opt => {
        if (opt.value === 'none') return;
        const excluded = isDbService && DB_EXCLUDED_PROVIDERS.includes(opt.dataset.provider);
        opt.hidden = excluded;
        opt.disabled = excluded;
    });
    const cur = selectEl.selectedOptions[0];
    if (isDbService && cur && DB_EXCLUDED_PROVIDERS.includes(cur.dataset.provider)) {
        selectEl.value = 'none';
        selectEl.dispatchEvent(new Event('change'));
    }
}

function loadProfileList() {
        fetch("/credentials")
            .then(response => {
                if (!response.ok) throw new Error(`HTTP ${response.status}`);
                return response.json();
            })
            .then(json => {
                const seen = new Set();
                const options = (Array.isArray(json) ? json : [])
                .filter((item) => {
                    // SQL Database / No-SQL 페이지는 RDB·NoSQL 인스턴스 관리를 지원하지 않는
                    // provider를 제외한다. Object Storage 페이지는 전체 노출
                    const dbPage = ['/generate/mysql', '/generate/no-sql', '/migrate/mysql', '/migrate/no-sql']
                        .some(p => window.location.pathname.includes(p));
                    if (dbPage && DB_EXCLUDED_PROVIDERS.includes(item.providerName)) return false;
                    // Object Storage 계열(비-DB) 화면에서는 azure/openstack 제외
                    if (!dbPage && OBJECTSTORAGE_EXCLUDED_PROVIDERS.includes(item.providerName)) return false;
                    if (seen.has(item.providerName)) return false;
                    seen.add(item.providerName);
                    return true;
                })
                .map((item) => {
                    return {
                        label: `${item.providerName.toUpperCase()} - ${item.credentialName}`,
                        value: item.providerName,
                        provider: item.providerName
                    }
                });

                // source/target 중 페이지에 있는 것만 채운다 (generate/backup은 한쪽만 있음).
                // credential이 1개뿐이면 자동 선택 + change를 쏴서 region/목록 로드를 태운다.
                ["sourceCredentialSelect", "targetCredentialSelect"].forEach(id => {
                    const selectEl = document.getElementById(id);
                    if (!selectEl) return;
                    options.forEach(optionData => {
                        const option = document.createElement("option");
                        option.value = optionData.value;
                        option.textContent = optionData.label;
                        option.setAttribute("data-provider", optionData.provider);
                        selectEl.appendChild(option);
                    });
                    if (options.length === 1) {
                        selectEl.selectedIndex = 1;
                        selectEl.dispatchEvent(new Event('change'));
                    }
                });

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
                    capOptions.forEach(optionData => {
                        const option = document.createElement("option");
                        option.value = optionData.value;
                        option.textContent = optionData.label;
                        capSelect.appendChild(option);
                    });
                }
            })
            .catch(reason => {
                console.error('loadProfileList:', reason);
                alert(reason);
            });
}

function migrationFormSubmit() {
    const form = document.getElementById('migForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const payload = new FormData(form);
        let jsonData = formDataToObject(payload)
        delete jsonData.sourceTime
        delete jsonData.targetTime
        delete jsonData.targetThread
        delete jsonData.targetTableCount
        delete jsonData.targetTableSize

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
        const service = document.getElementById('migService').value;
        jsonData.targetPoint.provider = getInputValue('targetPoint[provider]');
        jsonData.sourcePoint.provider = getInputValue('sourcePoint[provider]');
        jsonData.targetPoint.credentialId = parseInt(jsonData.targetPoint.credentialId);
        jsonData.sourcePoint.credentialId = parseInt(jsonData.sourcePoint.credentialId);

        if(service == "objectstorage") {
            applyFilter(jsonData)
        }

        let url = "/migrate/" + service;

        ensureEndpoint(jsonData.targetPoint);
        ensureEndpoint(jsonData.sourcePoint);

        TaskProgress.submit('migrate', url, jsonData);
    });

    TaskProgress.attach('migrate');   // 복귀 시 1회 조회로 복원
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
        if(service == "objectstorage" && (jsonData.sourcePoint.bucket == "none" || jsonData.sourcePoint.bucket == "-" || jsonData.sourcePoint.bucket == "")) {
            alert("please select bucket");
            return
        }        
        if(service == "objectstorage") {
            applyFilter(jsonData)
        } else {
            delete jsonData.sourceFilter;
        }
        loadingButtonOn();
        resultCollpase();

        let url = "/backup/" + service;

        jsonData.sourcePoint.credentialId = parseInt(jsonData.sourcePoint.credentialId);
        jsonData.sourcePoint.provider = provider

        if (service == "objectstorage") {
            ensureEndpoint(jsonData.sourcePoint);
        }

        TaskProgress.submit('backup', url, jsonData);
    });

    TaskProgress.attach('backup');   // 복귀 시 1회 조회로 복원
}

function RestoreFormSubmit() {
    const form = document.getElementById('restoreForm');

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const payload = new FormData(form);
        let jsonData = formDataToObject(payload)

        const provider = document.getElementById('targetPoint[provider]').value;
        const service = document.getElementById('targetService').value;
        if(service != "rdbms" && jsonData.targetPoint.credentialId == "none") {
            alert("credential not selected");
            return
        }
        if(jsonData.targetPoint.bucket == "none" || jsonData.targetPoint.bucket == "-" || jsonData.targetPoint.bucket == "") {
            alert("please select or create bucket");
            return
        }

        loadingButtonOn();
        resultCollpase();

        let url = "/restore/" + service;

        jsonData.targetPoint.credentialId = parseInt(jsonData.targetPoint.credentialId);
        jsonData.targetPoint.provider = provider       

        if(service == "objectstorage") {
            ensureEndpoint(jsonData.targetPoint);
        }

        TaskProgress.submit('restore', url, jsonData);
    });

    TaskProgress.attach('restore');   // 복귀 시 1회 조회로 복원
}

// 헤더 드롭다운의 서비스 생성/삭제 — confirm → 요청 → 결과 alert.
// clearServiceLink('/service/clearAll')는 header.html에서 링크가 주석 처리돼
// 도달 불가능하므로 핸들러를 두지 않는다. 링크를 되살리면 아래 배열에 추가할 것.
document.addEventListener('DOMContentLoaded', () => {
    [
        {
            id: 'genServiceLink', url: '/service/apply', method: 'POST',
            confirm: 'Are you sure you want to create a data-related service?',
            ok: 'Service creation request has been submitted.',
            fail: 'An error occurred while submitting the service creation request.',
        },
        {
            id: 'delServiceLink', url: '/service/destroy', method: 'DELETE',
            confirm: 'Are you sure you want to remove the data-related service?',
            ok: 'Service removal request has been submitted.',
            fail: 'An error occurred while submitting the service removal request.',
        },
    ].forEach(cfg => {
        const link = document.getElementById(cfg.id);
        if (!link) return;
        link.addEventListener('click', async event => {
            event.preventDefault();
            if (!confirm(cfg.confirm)) return;
            try {
                const res = await fetch(cfg.url, {
                    method: cfg.method,
                    headers: { 'Content-Type': 'application/json' },
                });
                if (!res.ok) {
                    const data = await res.json().catch(() => ({}));
                    throw new Error(data.message || cfg.fail);
                }
                alert(cfg.ok);
            } catch (error) {
                console.error(cfg.id + ':', error);
                alert(`Error: ${error.message}`);
            }
        });
    });
});

// m-cmp/mc-web-console이 iframe으로 감쌀 때 postMessage로 보내는 payload를 받는다.
// 페이로드에는 accessToken/workspaceInfo/projectInfo/operationId가 들어오지만
// 이 앱이 쓰는 건 projectInfo.ns_id 하나뿐이다 (→ 서버의 런타임 namespace로 세팅).
window.addEventListener("message", async function (event) {
    const data = event.data;
    try {
        const nsId = data?.projectInfo?.ns_id;
        if (!nsId) return;

        sessionStorage.setItem("nsId", nsId);

        await fetch("/namespace", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ nsId }),
        });
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

    const showMongo = service === "nrdbms" && (provider === "ncp" || provider === "alibaba");

    regionDiv.style.display = showMongo ? "none" : "";
    mongoDiv.style.display = showMongo ? "" : "none";

    const regionSelect = regionDiv.querySelector("select");
    if (regionSelect) {
        regionSelect.disabled = showMongo;
        if (showMongo) {
            regionSelect.removeAttribute('required');
        }
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
        //const regionSelect = document.getElementById(prefix + "RegionSelect");
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
            // provider가 바뀌면 이전 제출의 Result는 더 이상 유효하지 않으므로 비운다
            const resultText = document.getElementById('resultText');
            if (resultText) resultText.value = '';
            updateRegionsByProvider(prefix, provider);
            updateLabelByProvider(prefix, provider);
            if (provider !== 'none') {
                const regionSel = document.getElementById(prefix + "RegionSelect");
                if (regionSel && regionSel.value && regionSel.value !== 'none') {
                    regionSel.dispatchEvent(new Event('change'));
                }
            }
        });
    };

    // source/target 공통 처리
    ["source", "target"].forEach(initProviderHandlers);
});

function getServiceName(service, provider) {
    if (provider === "none") return "-";
  
    const serviceMap = {
      nrdbms: {
        aws: "AWS DynamoDB",
        ncp: "Naver MongoDB",
        gcp: "Google Firestore",
        alibaba: "Alibaba MongoDB",
      },
      objectstorage: {
        aws: "AWS S3",
        ncp: "Naver Object Storage",
        gcp: "Google Cloud Storage",
        alibaba: "Alibaba Object Storage",
        ibm: "IBM Cloud Object Storage",
        KT: "KT Cloud Object Storage",
        tencent: "Tencent Cloud Object Storage",
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

// ─────────────────────────────────────────────────────────────────────────────
// 목록 행 선택 체크박스(.row-select) 전역 동기화
//
// 모든 Bucket/Instance 목록의 선택 상태는 tr의 'table-active' 클래스로 표현된다.
// 체크박스는 pointer-events:none(styles.css)이라 클릭이 행으로 통과해 기존
// 행 클릭 핸들러가 선택을 처리하고, 여기서 클래스 변경을 관찰해 checked만 맞춘다.
// → 페이지별 동기화 코드 없이 단일 선택/자동 해제/재렌더가 항상 일치.
// ─────────────────────────────────────────────────────────────────────────────
new MutationObserver(muts => {
    muts.forEach(m => {
        if (m.target.tagName !== 'TR') return;
        const cb = m.target.querySelector('.row-select');
        if (cb) cb.checked = m.target.classList.contains('table-active');
    });
}).observe(document.body, { subtree: true, attributes: true, attributeFilter: ['class'] });

// ─────────────────────────────────────────────────────────────────────────────
// 공용 유틸 — 페이지별 중복 구현 금지
//
// 페이지 인라인 스크립트는 DOMContentLoaded 안에서 실행되므로(footer.html이
// scripts.js를 로드한 뒤 fire) 여기 정의된 전역을 안전하게 참조할 수 있다.
// ─────────────────────────────────────────────────────────────────────────────

// HTML 이스케이프. 속성값에 넣는 경우가 있어 큰따옴표까지 치환한다.
window.escHtml = function (s) {
    return String(s == null ? '—' : s)
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
};

// DB 인스턴스(rdbms/nrdbms) 상태 뱃지.
// 버킷 상태는 의미가 달라(생성 중 개념 없음) BucketPanel 내부본을 쓴다.
window.instanceStatusBadge = function (status) {
    const s = (status || '').toLowerCase();
    if (s === 'available') {
        return `<span class="text-success small"><i class="bi bi-check-circle-fill me-1"></i>available</span>`;
    }
    if (s === 'deleting') {
        return `<span class="text-danger small"><span class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>deleting</span>`;
    }
    if (s === 'failed') {
        return `<span class="text-danger small"><i class="bi bi-x-circle-fill me-1"></i>failed</span>`;
    }
    // 그 외 상태(net_creating 등)는 API가 반환한 값을 그대로 표시한다
    return `<span class="text-primary small"><i class="bi bi-hourglass-split ic-spin me-1"></i>${window.escHtml(s || '-')}</span>`;
};

// credential select가 채워질 때까지 대기한다.
// loadCredentialList()가 비동기로 option을 채우므로, 저장된 상태 복원 전에 기다려야 한다.
// 3초 내에 안 채워지면 그냥 진행한다(무한 대기 방지).
window.waitForCredentials = function (selectEl) {
    return new Promise(resolve => {
        if (!selectEl || selectEl.options.length > 1) { resolve(); return; }
        const obs = new MutationObserver(() => {
            if (selectEl.options.length > 1) { obs.disconnect(); resolve(); }
        });
        obs.observe(selectEl, { childList: true });
        setTimeout(() => { obs.disconnect(); resolve(); }, 3000);
    });
};

// 행 선택 시 Region 셀렉트를 인스턴스 region과 대소문자 무시로 동기화 (auto-select).
// rows에서 instanceId로 인스턴스를 찾아 그 region과 일치하는 option을 고른다.
//
// ⚠️ change 이벤트를 dispatch하지 않는다 — dispatch하면 목록이 리로드되어
//    방금 한 행 선택이 풀린다. 이 제약을 깨지 말 것.
window.autoSelectRegion = function (rows, instanceId, regionSel) {
    const inst = (rows || []).find(i => i.instanceId === instanceId);
    if (!inst || !inst.region || !regionSel) return;
    const opt = [...regionSel.options].find(o => (o.value || '').toLowerCase() === String(inst.region).toLowerCase());
    if (opt && regionSel.value !== opt.value) regionSel.value = opt.value;
};

// 페이지네이션 렌더 공용. container에 <nav>를 그리고 클릭 시 onPage(p)를 호출한다.
//
// ⚠️ totalPages는 '페이지 수'다 — 아이템 수를 넘기지 말 것.
//    아이템 수밖에 없으면 호출부에서 Math.ceil(items / PAGE_SIZE) || 1 로 변환한다.
// 페이지가 7개를 넘으면 생략부호(…)로 접고 이전/다음(‹ ›)을 붙인다.
window.renderPagination = function (container, totalPages, current, onPage) {
    if (!container) return;
    if (totalPages <= 1) { container.innerHTML = ''; return; }
    const range = [];
    if (totalPages <= 7) {
        for (let p = 1; p <= totalPages; p++) range.push(p);
    } else {
        range.push(1);
        if (current > 3) range.push('...');
        const s = Math.max(2, current - 1), e = Math.min(totalPages - 1, current + 1);
        for (let p = s; p <= e; p++) range.push(p);
        if (current < totalPages - 2) range.push('...');
        range.push(totalPages);
    }
    const prev = `<li class="page-item ${current === 1 ? 'disabled' : ''}"><a class="page-link" href="#" data-page="${current - 1}">‹</a></li>`;
    const next = `<li class="page-item ${current === totalPages ? 'disabled' : ''}"><a class="page-link" href="#" data-page="${current + 1}">›</a></li>`;
    const pages = range.map(p =>
        p === '...'
            ? `<li class="page-item disabled"><span class="page-link">…</span></li>`
            : `<li class="page-item ${p === current ? 'active' : ''}"><a class="page-link" href="#" data-page="${p}">${p}</a></li>`
    ).join('');
    container.innerHTML = `<nav><ul class="pagination pagination-sm mb-0 justify-content-center">${prev}${pages}${next}</ul></nav>`;
    container.querySelectorAll('a.page-link').forEach(link => {
        link.addEventListener('click', e => {
            e.preventDefault();
            const p = parseInt(link.dataset.page);
            if (!isNaN(p) && p >= 1 && p <= totalPages && p !== current) onPage(p);
        });
    });
};

// ─────────────────────────────────────────────────────────────────────────────
// BucketPanel — Object Storage 버킷 목록/탐색 공용 컴포넌트
//
// gen-object-storage.html의 버킷 테이블 + 가상 폴더 탐색 + 삭제 기능을
// migration/backup/restore 페이지에서 재사용하기 위한 모듈.
//
// 사용법:
//   페이지에 {{ template "bucket-panel.html" . }} 를 한 번 include 하고,
//   패널마다 아래 스켈레톤을 두고 init을 호출한다:
//     tbody id="{prefix}-body", div id="{prefix}-pagination", input id="{prefix}-search"
//   BucketPanel.init({
//     prefix: 'src-bkt',
//     cred: credSelectEl, region: regionSelectEl,
//     provider: providerHiddenEl, endpoint: endpointHiddenEl,
//     bucketInput: bucketFieldEl,   // 행 클릭 시 버킷명이 채워지는 input
//   })
//
// 특징:
// - provider 전체 버킷 표시 (리전 필터 없음, Region 컬럼으로 구분)
// - 탐색/삭제는 각 버킷의 소속 리전으로 호출 (선택된 리전과 무관)
// - Tumblebug가 prefix/delimiter를 지원하지 않아 flat 목록을 클라이언트에서
//   '/' 기준으로 폴더 그루핑 (maxKeys=1000 잘림은 길이 휴리스틱으로 배너 표시)
// ─────────────────────────────────────────────────────────────────────────────
const BucketPanel = (() => {
    const PAGE_SIZE = 10;

    // ── 공용 헬퍼 ────────────────────────────────────────
    const esc = window.escHtml;
    function fmtSize(bytes) {
        const n = Number(bytes);
        if (!isFinite(n) || n < 0) return '—';
        if (n < 1024) return n + ' B';
        if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB';
        if (n < 1024 * 1024 * 1024) return (n / (1024 * 1024)).toFixed(1) + ' MB';
        return (n / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    }
    function fmtDate(iso) {
        if (!iso) return '—';
        const d = new Date(iso);
        if (isNaN(d)) return esc(iso);
        const p = v => String(v).padStart(2, '0');
        return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
    }
    function statusBadge(status) {
        const s = (status || '').toLowerCase();
        if (s === 'available') {
            return `<span class="text-success small"><i class="bi bi-check-circle-fill me-1"></i>available</span>`;
        }
        if (!s) return `<span class="text-muted small">—</span>`;
        return `<span class="text-secondary small">${esc(status)}</span>`;
    }
    // ── 패널별 targetPoint 구성 (provider별 endpoint 규칙 포함) ──
    function buildTp(panel, extra) {
        const provider = panel.provider.value || '';
        const tp = Object.assign({ provider, region: panel.region.value }, extra || {});
        if (provider === 'ncp') {
            tp.endpoint = tp.endpoint || panel.endpoint?.value || NCP_OBJECTSTORAGE_ENDPOINT;
        } else if (provider === 'alibaba') {
            tp.endpoint = tp.endpoint || panel.endpoint?.value || resolveAlibabaEndpoint(tp.region);
        }
        return tp;
    }
    function bucketRegionOf(b, provider) {
        return b.connectionConfig?.regionZoneInfo?.assignedRegion
            || (b.connectionName || '').replace(`${provider}-`, '')
            || '';
    }

    // Created 컬럼: conditions에서 type === 'Ready'인 조건의 lastTransitionTime만 사용
    function bucketReadyTime(b) {
        const c = (b.conditions || []).find(x => x.type === 'Ready');
        return c ? c.lastTransitionTime : '';
    }

    // ── 탐색 Offcanvas (페이지당 싱글턴, bucket-panel.html 파셜) ──
    let browse = null; // { panel, name, region } — 오브젝트 탐색 중인 버킷
    let deleteTarget = null; // { panel, name, region } — 삭제 확인 모달 대상 (탐색과 독립: 행에서 직접 삭제 가능)
    let createTarget = null; // { panel } — 생성 모달을 연 패널 (mig처럼 패널이 둘일 때 컨텍스트 구분)
    let allObjects = [], objectPage = 1, currentPrefix = '', objectsTruncated = false;
    let pendingDeleteKey = null;
    let ocRefs = null; // lazy: 파셜 DOM + bootstrap 인스턴스

    function ocInit() {
        if (ocRefs) return ocRefs;
        const el = document.getElementById('bktOffcanvas');
        if (!el) { console.warn('BucketPanel: bucket-panel.html 파셜이 include되지 않았습니다'); return null; }
        ocRefs = {
            el,
            oc: new bootstrap.Offcanvas(el, { backdrop: true, scroll: false }),
            title: document.getElementById('bkt-oc-title'),
            count: document.getElementById('bkt-object-count'),
            truncated: document.getElementById('bkt-object-truncated'),
            breadcrumb: document.getElementById('bkt-object-breadcrumb'),
            tbody: document.getElementById('bkt-object-body'),
            pagination: document.getElementById('bkt-object-pagination'),
            deleteModal: new bootstrap.Modal(document.getElementById('bktDeleteModal')),
            objDeleteModal: new bootstrap.Modal(document.getElementById('bktObjectDeleteModal')),
            // 생성/탐색 2모드 (gen-object-storage와 동일 구조)
            context: document.getElementById('bkt-oc-context'),
            sectionCreate: document.getElementById('bkt-section-create'),
            sectionObjects: document.getElementById('bkt-section-objects'),
            createInput: document.getElementById('bkt-create-input'),
            createConfirm: document.getElementById('bkt-create-confirm'),
        };
        if (ocRefs.sectionCreate) {
            ocRefs.createInput.addEventListener('input', function () {
                const name = this.value.trim();
                let msg = bucketNameError(name);
                if (!msg && name && createTarget?.panel.allBuckets.some(b => b.name === name)) {
                    msg = 'This bucket name already exists.';
                }
                this.classList.toggle('is-invalid', !!msg);
                this.closest('.form-floating')?.classList.toggle('is-invalid', !!msg);
                const errEl = document.getElementById('bkt-create-error');
                errEl.textContent = msg;
                errEl.classList.toggle('d-none', !msg);
                ocRefs.createConfirm.disabled = !name || !!msg;
            });
            ocRefs.createConfirm.addEventListener('click', createBucket);
        }
        document.getElementById('bkt-object-refresh').addEventListener('click', loadObjects);
        document.getElementById('bktDeleteModal').addEventListener('show.bs.modal', () => {
            document.getElementById('bkt-delete-input').value = '';
            document.getElementById('bkt-delete-error').classList.add('d-none');
            document.getElementById('bkt-delete-confirm').disabled = true;
        });
        document.getElementById('bkt-delete-input').addEventListener('input', function () {
            if (!deleteTarget) return;
            const match = this.value === deleteTarget.name;
            document.getElementById('bkt-delete-error').classList.toggle('d-none', !this.value || match);
            document.getElementById('bkt-delete-confirm').disabled = !match;
        });
        document.getElementById('bkt-delete-confirm').addEventListener('click', deleteBucket);
        document.getElementById('bkt-object-delete-confirm').addEventListener('click', deleteObject);
        return ocRefs;
    }

    // ── 버킷 생성 (gen-object-storage와 동일 규칙/API) ─────────
    function bucketNameError(name) {
        if (!name) return '';
        if (name.length < 3 || name.length > 63) return 'Bucket name must be 3–63 characters.';
        if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(name)) return 'Only lowercase letters, numbers, and hyphens (-) are allowed, and it cannot start or end with a hyphen.';
        return '';
    }

    // 패널의 credential/region으로 offcanvas 컨텍스트 문구 갱신
    function updateOcContext(panel) {
        const credText = panel.cred?.options?.[panel.cred.selectedIndex]?.textContent?.trim() || panel.provider.value || '—';
        const region = panel.region?.value;
        ocRefs.context.textContent = (region && region !== 'none') ? `${credText} / ${region}` : credText;
    }

    // 생성 모드로 offcanvas 열기 (gen-object-storage enterCreateMode와 동일 UX)
    function enterCreateMode(panel) {
        const r = ocInit();
        if (!r || !r.sectionCreate) return;
        createTarget = { panel };
        r.title.innerHTML = '<i class="bi bi-plus-circle me-2 text-primary"></i>Create Bucket';
        r.sectionCreate.classList.remove('d-none');
        r.sectionObjects.classList.add('d-none');
        r.createConfirm.classList.remove('d-none');
        r.createInput.value = '';
        r.createInput.classList.remove('is-invalid');
        r.createInput.closest('.form-floating')?.classList.remove('is-invalid');
        document.getElementById('bkt-create-error').classList.add('d-none');
        r.createConfirm.disabled = true;
        updateOcContext(panel);
        r.oc.show();
    }

    async function createBucket() {
        if (!createTarget) return;
        const panel = createTarget.panel;
        const name = ocRefs.createInput.value.trim();
        if (!name || bucketNameError(name)) return;
        const btn = ocRefs.createConfirm;
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>Creating...';
        try {
            const res = await fetch('/objectstorage/buckets', {
                method: 'PUT', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ targetPoint: buildTp(panel, { bucket: name }) })
            });
            const json = await res.json().catch(() => ({}));
            if (!res.ok) throw new Error(json.Result || `HTTP ${res.status}`);
            if ((json.Result || '').includes('already exists')) {
                alert(`Bucket '${name}' already exists.`);
            }
            ocRefs.oc.hide();
            // 방금 만든 버킷을 바로 선택 상태로 (gen-object-storage와 동일 UX)
            // select 페이지는 목록 갱신으로 옵션이 생긴 뒤에 값을 넣어야 반영된다
            await panel.load();
            if (panel.bucketInput) {
                panel.bucketInput.value = name;
                panel.selectedRegion = panel.region.value || null;
                panel.bucketInput.dispatchEvent(new Event('input'));
            }
        } catch (err) {
            alert('Failed to create bucket: ' + err.message);
        } finally {
            btn.disabled = false;
            btn.textContent = 'Create';
        }
    }

    // 버킷 삭제 확인 모달 열기 — 유일한 진입점: 각 패널 툴바의 Delete 버튼 (선택된 버킷 대상)
    function openDeleteModal(panel, name, region) {
        const r = ocInit();
        if (!r) return;
        deleteTarget = { panel, name, region: region || panel.region.value };
        document.getElementById('bkt-delete-name').textContent = name;
        r.deleteModal.show();
    }

    function openBrowse(panel, name, region) {
        const r = ocInit();
        if (!r) return;
        browse = { panel, name, region: region || panel.region.value };
        currentPrefix = '';
        r.title.innerHTML = `<i class="bi bi-bucket me-2 text-primary"></i>${esc(name)}`;
        if (r.sectionCreate) {
            r.sectionCreate.classList.add('d-none');
            r.sectionObjects.classList.remove('d-none');
            r.createConfirm.classList.add('d-none');
            updateOcContext(panel);
        }
        r.oc.show();
        loadObjects();
    }

    async function loadObjects() {
        if (!browse) return;
        const r = ocRefs;
        r.tbody.innerHTML = '<tr><td colspan="4" class="text-center py-3"><span class="spinner-border spinner-border-sm text-muted"></span></td></tr>';
        try {
            const res = await fetch('/objectstorage/buckets/objects', {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ targetPoint: buildTp(browse.panel, { bucket: browse.name, region: browse.region }) })
            });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const json = await res.json();
            allObjects = json.objects || [];
            // ponytail: Tumblebug maxKeys=1000 고정 — 길이로 잘림 추정
            objectsTruncated = allObjects.length >= 1000;
            objectPage = 1;
            renderObjects();
        } catch (e) {
            console.error('BucketPanel.loadObjects:', e);
            r.tbody.innerHTML = '<tr><td colspan="4" class="text-center text-danger small py-3">Failed to load object list</td></tr>';
            r.pagination.innerHTML = '';
        }
    }

    function levelEntries() {
        const folders = new Map();
        const files = [];
        for (const o of allObjects) {
            const key = o.key || '';
            if (!key.startsWith(currentPrefix)) continue;
            const rest = key.slice(currentPrefix.length);
            if (!rest) continue;
            const idx = rest.indexOf('/');
            if (idx >= 0) {
                const dir = rest.slice(0, idx);
                folders.set(dir, (folders.get(dir) || 0) + 1);
            } else {
                files.push(o);
            }
        }
        const folderRows = [...folders.entries()]
            .sort((a, b) => a[0].localeCompare(b[0]))
            .map(([name, count]) => ({ type: 'folder', name, count }));
        files.sort((a, b) => (a.key || '').localeCompare(b.key || ''));
        return folderRows.concat(files.map(o => ({ type: 'file', o })));
    }

    function renderBreadcrumb() {
        const ol = ocRefs.breadcrumb;
        const segs = currentPrefix ? currentPrefix.split('/').filter(Boolean) : [];
        let html = `<li class="breadcrumb-item${segs.length ? '' : ' active'}">` +
            (segs.length ? `<a href="#" data-prefix="">${esc(browse.name)}</a>` : esc(browse.name)) + '</li>';
        let acc = '';
        segs.forEach((s, i) => {
            acc += s + '/';
            const isLast = i === segs.length - 1;
            html += `<li class="breadcrumb-item${isLast ? ' active' : ''}">` +
                (isLast ? esc(s) : `<a href="#" data-prefix="${esc(acc)}">${esc(s)}</a>`) + '</li>';
        });
        ol.innerHTML = html;
        ol.querySelectorAll('a').forEach(a => {
            a.addEventListener('click', e => {
                e.preventDefault();
                currentPrefix = a.dataset.prefix || '';
                objectPage = 1;
                renderObjects();
            });
        });
    }

    function renderObjects() {
        const r = ocRefs;
        r.count.textContent = allObjects.length;
        r.truncated.classList.toggle('d-none', !objectsTruncated);
        renderBreadcrumb();

        const entries = levelEntries();
        const totalPages = Math.max(1, Math.ceil(entries.length / PAGE_SIZE));
        if (objectPage > totalPages) objectPage = totalPages;

        if (!entries.length) {
            r.tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted small py-3">No objects found</td></tr>';
        } else {
            const start = (objectPage - 1) * PAGE_SIZE;
            r.tbody.innerHTML = entries.slice(start, start + PAGE_SIZE).map(en => {
                if (en.type === 'folder') {
                    return `<tr class="bkt-folder-row" style="cursor:pointer;" data-dir="${esc(en.name)}">
                        <td class="small"><i class="bi bi-folder-fill text-warning me-2"></i>${esc(en.name)}/</td>
                        <td class="small text-muted">${en.count} item${en.count > 1 ? 's' : ''}</td>
                        <td class="small text-muted">—</td>
                        <td></td>
                    </tr>`;
                }
                const o = en.o;
                const leaf = (o.key || '').slice(currentPrefix.length);
                return `<tr>
                    <td class="small text-truncate" style="max-width:280px;" title="${esc(o.key)}"><i class="bi bi-file-earmark me-2 text-muted"></i>${esc(leaf)}</td>
                    <td class="small">${fmtSize(o.size)}</td>
                    <td class="small">${fmtDate(o.lastModified)}</td>
                    <td class="text-center">
                        <button type="button" class="btn btn-link btn-sm text-danger p-0 bkt-object-delete-btn" data-key="${esc(o.key)}" title="Delete">
                            <i class="bi bi-trash3"></i>
                        </button>
                    </td>
                </tr>`;
            }).join('');
            r.tbody.querySelectorAll('tr.bkt-folder-row').forEach(row => {
                row.addEventListener('click', () => {
                    currentPrefix += row.dataset.dir + '/';
                    objectPage = 1;
                    renderObjects();
                });
            });
            r.tbody.querySelectorAll('.bkt-object-delete-btn').forEach(btn => {
                btn.addEventListener('click', e => {
                    e.stopPropagation();
                    pendingDeleteKey = btn.dataset.key;
                    document.getElementById('bkt-object-delete-key').textContent = pendingDeleteKey;
                    ocRefs.objDeleteModal.show();
                });
            });
        }
        renderPagination(r.pagination, totalPages, objectPage, p => { objectPage = p; renderObjects(); });
    }

    async function deleteObject() {
        if (!browse || !pendingDeleteKey) return;
        const btn = document.getElementById('bkt-object-delete-confirm');
        const key = pendingDeleteKey;
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>Deleting...';
        try {
            const res = await fetch('/objectstorage/buckets/object', {
                method: 'DELETE', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ targetPoint: buildTp(browse.panel, { bucket: browse.name, region: browse.region }), objectKey: key })
            });
            if (!res.ok) {
                const json = await res.json().catch(() => ({}));
                throw new Error(json.Result || `HTTP ${res.status}`);
            }
            ocRefs.objDeleteModal.hide();
            allObjects = allObjects.filter(o => o.key !== key);
            renderObjects();
        } catch (err) {
            alert('Failed to delete object: ' + err.message);
        } finally {
            pendingDeleteKey = null;
            btn.disabled = false;
            btn.innerHTML = '<i class="bi bi-trash3 me-1"></i>Delete';
        }
    }

    async function deleteBucket() {
        if (!deleteTarget) return;
        const { panel, name, region } = deleteTarget;
        const btn = document.getElementById('bkt-delete-confirm');
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>Deleting...';
        try {
            const res = await fetch('/objectstorage/buckets', {
                method: 'DELETE', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ targetPoint: buildTp(panel, { bucket: name, region }) })
            });
            if (!res.ok) {
                const json = await res.json().catch(() => ({}));
                throw new Error(json.Result || `HTTP ${res.status}`);
            }
            ocRefs.deleteModal.hide();
            // 탐색 중이던 버킷을 삭제했으면 offcanvas도 닫는다
            if (browse && browse.name === name) { ocRefs.oc.hide(); browse = null; }
            if (panel.bucketInput && panel.bucketInput.value.trim() === name) {
                panel.bucketInput.value = '';
                panel.bucketInput.dispatchEvent(new Event('input'));
            }
            deleteTarget = null;
            panel.load();
        } catch (err) {
            alert('Failed to delete bucket: ' + err.message);
        } finally {
            btn.disabled = false;
            btn.innerHTML = '<i class="bi bi-trash3 me-1"></i>Delete';
        }
    }

    // ── 패널(버킷 테이블) ────────────────────────────────
    function init(cfg) {
        const panel = {
            prefix: cfg.prefix,
            cred: cfg.cred, region: cfg.region,
            provider: cfg.provider, endpoint: cfg.endpoint,
            bucketInput: cfg.bucketInput,
            deleteBtn: cfg.deleteBtn || null, // 버킷 삭제의 유일한 진입점 (목록 툴바 Delete 버튼)
            selectedRegion: null,             // 선택된 버킷의 실제 리전 (Delete 시 사용)
            tbody: document.getElementById(cfg.prefix + '-body'),
            pagination: document.getElementById(cfg.prefix + '-pagination'),
            search: document.getElementById(cfg.prefix + '-search'),
            allBuckets: [], page: 1, query: '',
            load, clear,
        };

        // 선택 상태에 따라 툴바 Delete 버튼 활성화 (생성 중 버킷은 삭제 불가)
        function updateDeleteBtn() {
            if (!panel.deleteBtn) return;
            const name = (panel.bucketInput?.value || '').trim();
            const b = name ? panel.allBuckets.find(x => x.name === name) : null;
            const isCreating = !!b && (b.status || '').toLowerCase().includes('creating');
            panel.deleteBtn.disabled = !name || isCreating;
        }

        function syncHighlight() {
            const name = (panel.bucketInput?.value || '').trim();
            panel.tbody.querySelectorAll('tr.bkt-row').forEach(r =>
                r.classList.toggle('table-active', r.dataset.name === name));
        }

        // bucketInput이 <select>면 로드된 버킷 이름으로 옵션 구성 (Target Bucket Select 페이지)
        function populateBucketSelect() {
            const sel = panel.bucketInput;
            if (!sel || sel.tagName !== 'SELECT') return;
            const cur = sel.value;
            sel.replaceChildren(new Option('- Select Bucket -', ''));
            [...new Set(panel.allBuckets.map(b => b.name))].forEach(n => sel.add(new Option(n, n)));
            sel.value = cur && panel.allBuckets.some(b => b.name === cur) ? cur : '';
        }

        function render() {
            const q = panel.query.trim().toLowerCase();
            const filtered = q ? panel.allBuckets.filter(b => (b.name || '').toLowerCase().includes(q)) : panel.allBuckets;
            const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
            if (panel.page > totalPages) panel.page = totalPages;

            if (!filtered.length) {
                panel.tbody.innerHTML = '<tr><td colspan="7" class="text-center text-muted small py-3">No buckets found</td></tr>';
            } else {
                const provider = panel.provider.value || '';
                const selected = (panel.bucketInput?.value || '').trim();
                const start = (panel.page - 1) * PAGE_SIZE;
                panel.tbody.innerHTML = filtered.slice(start, start + PAGE_SIZE).map(b => {
                    const region = bucketRegionOf(b, provider);
                    // connectionName은 항상 '{provider}-{region}'이라 옵션값과 일치한다 (NCP 매칭용)
                    const connRegion = (b.connectionName || '').replace(`${provider}-`, '');
                    // 표시용 리전명은 regionDetail.regionName, 없으면 기존 파생 리전값
                    const regionName = b.connectionConfig?.regionDetail?.regionName || region;
                    const isSel = selected === b.name;
                    return `<tr class="bkt-row${isSel ? ' table-active' : ''}" style="cursor:pointer;"
                            data-name="${esc(b.name)}" data-region="${esc(region)}" data-conn-region="${esc(connRegion)}">
                        <td class="text-center align-middle"><input type="checkbox" class="form-check-input row-select"${isSel ? ' checked' : ''}></td>
                        <td class="small text-truncate" style="max-width:180px;" title="${esc(b.cspResourceId || '')}">${esc(b.cspResourceId || '-')}</td>
                        <td class="small">${esc(b.name)}</td>
                        <td class="small">${esc(regionName)}</td>
                        <td>${statusBadge(b.status)}</td>
                        <td class="small text-nowrap">${esc(formatKstDate(bucketReadyTime(b)))}</td>
                        <td class="text-center text-nowrap">
                            <button type="button" class="btn btn-link btn-sm p-0 bkt-browse-btn" data-name="${esc(b.name)}" data-region="${esc(region)}" title="Browse objects">
                                <i class="bi bi-folder2-open"></i>
                            </button>
                        </td>
                    </tr>`;
                }).join('');
                panel.tbody.querySelectorAll('tr.bkt-row').forEach(row => {
                    row.addEventListener('click', () => {
                        if (!panel.bucketInput) return;
                        const name = row.dataset.name;
                        if (panel.bucketInput.value.trim() === name) {
                            panel.bucketInput.value = '';
                            panel.selectedRegion = null;
                        } else {
                            panel.bucketInput.value = name;
                            panel.selectedRegion = row.dataset.region || null;
                            // 버킷 실제 리전으로 Region 셀렉트를 조용히 동기화 (대소문자 무시 매칭).
                            // assignedRegion(예: NCP 'KR')이 옵션과 형식이 다르면 connectionName 파생값(kr-standard)으로 매칭.
                            // change는 발생시키지 않는다 → 목록 리로드로 선택이 풀리는 것을 방지.
                            const matched = matchRegionOption(panel.region, row.dataset.region, row.dataset.connRegion);
                            if (matched) {
                                if (panel.region.value !== matched) panel.region.value = matched;
                                panel.selectedRegion = matched;
                            }
                        }
                        panel.bucketInput.dispatchEvent(new Event('input'));
                        syncHighlight();
                        updateDeleteBtn();
                    });
                });
                panel.tbody.querySelectorAll('.bkt-browse-btn').forEach(btn => {
                    btn.addEventListener('click', e => {
                        e.stopPropagation();
                        openBrowse(panel, btn.dataset.name, btn.dataset.region);
                    });
                });
            }
            renderPagination(panel.pagination, totalPages, panel.page, p => { panel.page = p; render(); });
        }

        async function load() {
            const provider = panel.provider.value || '';
            if (!provider || provider === 'none') {
                panel.allBuckets = [];
                populateBucketSelect();
                panel.tbody.innerHTML = '<tr><td colspan="7" class="text-center text-muted small py-3">Select a credential to list buckets</td></tr>';
                panel.pagination.innerHTML = '';
                return;
            }
            panel.tbody.innerHTML = '<tr><td colspan="7" class="text-center py-3"><span class="spinner-border spinner-border-sm text-muted"></span></td></tr>';
            try {
                const res = await fetch(`/objectstorage/buckets?filterKey=providerName&filterVal=${encodeURIComponent(provider)}`, {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ targetPoint: buildTp(panel) })
                });
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                const json = await res.json();
                panel.allBuckets = (json.objectStorage || []).filter(b =>
                    (b.connectionConfig?.providerName || b.connectionName || '').toLowerCase().startsWith(provider.toLowerCase()));
                panel.page = 1;
                populateBucketSelect();
                render();
                updateDeleteBtn(); // 재조회로 상태가 바뀐 선택 버킷(creating→available)의 Delete 재평가
            } catch (e) {
                console.error('BucketPanel.load:', e);
                panel.tbody.innerHTML = '<tr><td colspan="7" class="text-center text-danger small py-3">Failed to load bucket list</td></tr>';
                panel.pagination.innerHTML = '';
            }
        }

        function clear() {
            if (panel.bucketInput) {
                panel.bucketInput.value = '';
                panel.bucketInput.dispatchEvent(new Event('input'));
            }
            panel.selectedRegion = null;
            syncHighlight();
            updateDeleteBtn();
        }

        if (panel.search) {
            panel.search.addEventListener('input', function () {
                panel.query = this.value;
                panel.page = 1;
                render();
            });
        }
        if (panel.bucketInput) {
            panel.bucketInput.addEventListener('input', () => { syncHighlight(); updateDeleteBtn(); });
        }
        // Target Bucket <select>: 선택 시 버킷 실제 리전 동기화 (행 클릭과 동일 규칙)
        if (panel.bucketInput?.tagName === 'SELECT') {
            panel.bucketInput.addEventListener('change', () => {
                const provider = panel.provider.value || '';
                const b = panel.allBuckets.find(x => x.name === panel.bucketInput.value);
                panel.selectedRegion = null;
                if (b) {
                    const region = bucketRegionOf(b, provider);
                    const connRegion = (b.connectionName || '').replace(`${provider}-`, '');
                    const matched = matchRegionOption(panel.region, region, connRegion);
                    panel.selectedRegion = matched || region || null;
                    if (matched && panel.region.value !== matched) panel.region.value = matched;
                }
                panel.bucketInput.dispatchEvent(new Event('input')); // 하이라이트/버튼 동기화 재사용
            });
        }
        // 툴바 Delete 버튼: 선택된 버킷 삭제 (유일한 삭제 진입점)
        if (panel.deleteBtn) {
            panel.deleteBtn.addEventListener('click', () => {
                const name = (panel.bucketInput?.value || '').trim();
                if (!name) return;
                openDeleteModal(panel, name, panel.selectedRegion || panel.region.value);
            });
            updateDeleteBtn();
        }
        // 툴바 Create(+ Bucket) 버튼: credential 선택 시에만 활성 (옵트인 — 전달한 페이지만)
        if (cfg.createBtn) {
            panel.createBtn = cfg.createBtn;
            const updateCreateBtn = () => {
                const p = panel.provider.value || '';
                panel.createBtn.disabled = !p || p === 'none';
            };
            panel.createBtn.addEventListener('click', () => enterCreateMode(panel));
            panel.provider.addEventListener('change', updateCreateBtn);
            updateCreateBtn();
        }
        // provider(=credential) 변경 시 자동 재조회. 페이지 스크립트가
        // initProviderHandlers를 통해 provider hidden change를 dispatch한다
        panel.provider.addEventListener('change', () => { clear(); load(); });

        return panel;
    }

    return { init };
})();

// ─────────────────────────────────────────────────────────────────────────────
// InstancePanel — DB 인스턴스 생성/삭제 공용 컴포넌트 (SQL·NoSQL)
//
// gen-mysql/gen-no-sql의 생성·삭제 기능을 migration/backup/restore 페이지에서
// 재사용하기 위한 모듈. 목록 조회/렌더는 각 페이지의 기존 구현을 그대로 두고,
// 이 모듈은 생성 Offcanvas + 삭제 모달 + 완료 폴링만 담당한다.
//
// 사용법:
//   페이지에 {{ template "instance-panel.html" . }} 를 한 번 include 하고:
//   const panel = InstancePanel.init({
//     kind: 'rdbms' | 'nrdbms',            // API: /db/{kind}
//     getProvider: () => str, getRegion: () => str,
//     getCredLabel: () => str,             // offcanvas 컨텍스트 표기용 (옵션)
//     createBtn: el, deleteBtn: el,        // 각 목록 툴바의 버튼
//     getSelected: () => ({instanceId, name}) | null,
//     onData: (instances) => void,         // 폴링/갱신 결과를 페이지 목록에 반영
//   });
//   행 선택이 바뀔 때 panel.refreshButtons() 를 호출한다.
// ─────────────────────────────────────────────────────────────────────────────
const InstancePanel = (() => {
    const POLL_MS = 5000;

    // 생성 직후 첫 조회를 늦출 provider (ms). alibaba는 결과적 일관성 때문에
    // CreateInstance 응답 직후 DescribeDBInstances가 새 인스턴스를 아직 반환하지 않는다
    // → 첫 틱이 즉시 돌면 "생성했는데 목록 그대로"로 보인다. 첫 틱만 늦추고 이후는 POLL_MS 그대로.
    // 값은 실측이 아닌 추정치 — 체감이 맞지 않으면 이 숫자만 조정한다.
    const CREATE_SETTLE_MS = { alibaba: 5000 };

    // ── provider별 계정/비밀번호 규칙 (gen-mysql/gen-no-sql과 동일 기준) ──
    // Alibaba: 클라이언트에서는 admin/root/user/test만 금지 (gen-mysql과 동일) —
    // 실제 예약어 위반은 생성 시 CSP API 응답(InvalidAccountName.Forbid)으로 안내된다.
    const RDB_ALIBABA_BLOCKED = ['admin', 'root', 'user', 'test'];
    const RDB_NCP_BLOCKED = ['root','admin','radmin','agent','api_admin','ha_admin',
        'ragent','repl_admin','mysql.session','mysql.sys'];
    const COMMON_PW = ['password','password1','password123','12345678','123456789',
        'qwerty123','p@ssw0rd','passw0rd','welcome1','admin123','admin1234','letmein1','iloveyou1'];

    function usernameHelp(kind, provider) {
        const p = (provider || '').toLowerCase();
        if (kind === 'nrdbms') {
            if (p === 'alibaba') return "Only 'root' is allowed.";
            if (p === 'ncp') return "'admin' and 'root' are not allowed.";
            return '';
        }
        if (p === 'gcp') return "Only 'root' is allowed.";
        if (p === 'alibaba') return 'Not allowed: admin, root, user, test.';
        if (p === 'ncp') return 'Not allowed: root, admin, radmin, agent, api_admin, ha_admin, ragent, repl_admin, mysql.session, mysql.sys.';
        return '';
    }
    function usernameError(kind, provider, val) {
        const v = (val || '').trim();
        if (!v) return '';
        const p = (provider || '').toLowerCase();
        const lc = v.toLowerCase();
        if (kind === 'nrdbms') {
            if (p === 'alibaba' && lc !== 'root') return 'Only "root" is allowed for Alibaba.';
            if (p === 'ncp' && (lc === 'admin' || lc === 'root')) return '"admin" and "root" are not allowed for NCP.';
            return '';
        }
        if (p === 'alibaba') {
            if (v.length > 32) return 'Username must not exceed 32 characters.';
            if (!/^[a-zA-Z0-9_]+$/.test(v)) return 'Username can only contain letters, numbers, and underscores (_).';
            if (!/^[a-zA-Z]/.test(v)) return 'Username must start with a letter.';
            if (!/[a-zA-Z0-9]$/.test(v)) return 'Username must end with a letter or number.';
            if (RDB_ALIBABA_BLOCKED.includes(lc)) return `'${v}' is not allowed as a username.`;
        }
        if (p === 'ncp' && RDB_NCP_BLOCKED.includes(lc)) return `'${v}' is not allowed as a username.`;
        // gen-mysql과 동일: GCP는 root만 허용 (강제 채움 대신 검증+안내)
        if (p === 'gcp' && lc !== 'root') return "Only 'root' is allowed.";
        return '';
    }
    function passwordError(provider, val) {
        if (!val) return '';
        if ((provider || '').toLowerCase() !== 'alibaba') return '';
        if (val.length < 8 || val.length > 32) return 'Password must be 8–32 characters.';
        const cats = [/[A-Z]/.test(val), /[a-z]/.test(val), /[0-9]/.test(val), /[!@#$%^&*()_+\-=]/.test(val)];
        if (cats.filter(Boolean).length < 3) return 'Include at least 3 of: uppercase, lowercase, digit, special character (!@#$%^&*()_+-=).';
        if (COMMON_PW.includes(val.toLowerCase())) return 'This password is too common. Please choose a stronger password.';
        return '';
    }
    // instance class를 family/size로 분리 (gen-mysql icSplit과 동일 기준)
    function icSplit(cls, provider) {
        const p = (provider || '').toLowerCase();
        if (p === 'gcp') {
            const parts = cls.split('-');
            return { family: parts.slice(0, 2).join('-'), size: parts.slice(2).join('-') };
        }
        const n = p === 'ncp' ? 4 : 2;
        const parts = cls.split('.');
        return { family: parts.slice(0, n).join('.'), size: parts.slice(n).join('.') };
    }

    // ── 파셜 refs (lazy 싱글턴) + 활성 패널 ──────────────────
    let refs = null;
    let active = null;      // 오프캔버스를 연 패널
    let deleteCtx = null;   // { panel, instanceId, name }
    let engineVersionData = [];
    let icClasses = [];     // 현재 Instance Class 목록 (2패널 드롭다운)
    let pollTimer = null;

    function uiInit() {
        if (refs) return refs;
        const oc = document.getElementById('dbiOffcanvas');
        if (!oc) { console.warn('InstancePanel: instance-panel.html 파셜이 include되지 않았습니다'); return null; }
        refs = {
            oc: new bootstrap.Offcanvas(oc, { backdrop: true, scroll: false }),
            deleteModal: new bootstrap.Modal(document.getElementById('dbiDeleteModal')),
            engine: document.getElementById('dbi-engine'),
            version: document.getElementById('dbi-engineVersion'),
            klass: document.getElementById('dbi-instanceClass'),
            icWrapper: document.getElementById('dbi-icWrapper'),
            icTrigger: document.getElementById('dbi-icTrigger'),
            icTriggerText: document.getElementById('dbi-icTriggerText'),
            icPanel: document.getElementById('dbi-icPanel'),
            icFamilyList: document.getElementById('dbi-icFamilyList'),
            icSizeList: document.getElementById('dbi-icSizeList'),
            instanceId: document.getElementById('dbi-instanceId'),
            storage: document.getElementById('dbi-storage'),
            username: document.getElementById('dbi-username'),
            password: document.getElementById('dbi-password'),
            createBtn: document.getElementById('dbi-createBtn'),
        };
        refs.engine.addEventListener('change', () => { populateVersions(); updateCreateBtn(); });
        refs.version.addEventListener('change', () => { loadClasses(); updateCreateBtn(); });
        refs.icTrigger.addEventListener('click', () => {
            if (!icClasses.length) return;
            refs.icPanel.classList.toggle('d-none');
        });
        refs.icTrigger.addEventListener('keydown', e => {
            if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); refs.icTrigger.click(); }
            if (e.key === 'Escape') icClose();
        });
        document.addEventListener('click', e => {
            if (!refs.icWrapper.contains(e.target)) icClose();
        });
        // 중복 Instance ID: is-invalid + invalid-feedback로 표시 (storage와 동일 패턴)
        refs.instanceId.addEventListener('input', () => {
            const dup = instanceIdDup();
            refs.instanceId.classList.toggle('is-invalid', dup);
            refs.instanceId.closest('.form-floating').classList.toggle('is-invalid', dup);
            updateCreateBtn();
        });
        // gen-mysql/gen-no-sql과 동일: is-invalid + invalid-feedback로 표시
        refs.storage.addEventListener('input', () => {
            const min = Number(refs.storage.min) || 20;
            const bad = refs.storage.value !== '' && Number(refs.storage.value) < min;
            refs.storage.classList.toggle('is-invalid', bad);
            refs.storage.closest('.form-floating').classList.toggle('is-invalid', bad);
            updateCreateBtn();
        });
        refs.username.addEventListener('input', () => {
            if (!active) return;
            const msg = usernameError(active.kind, active.getProvider(), refs.username.value);
            const el = document.getElementById('dbi-username-error');
            el.textContent = msg;
            el.classList.toggle('d-none', !msg);
            updateCreateBtn();
        });
        refs.password.addEventListener('input', () => {
            if (!active) return;
            const msg = passwordError(active.getProvider(), refs.password.value);
            const el = document.getElementById('dbi-password-error');
            el.textContent = msg;
            el.classList.toggle('d-none', !msg);
            updateCreateBtn();
        });
        refs.createBtn.addEventListener('click', createInstance);
        document.getElementById('dbiDeleteModal').addEventListener('show.bs.modal', () => {
            document.getElementById('dbi-delete-input').value = '';
            document.getElementById('dbi-delete-error').classList.add('d-none');
            document.getElementById('dbi-delete-confirm').disabled = true;
        });
        document.getElementById('dbi-delete-input').addEventListener('input', function () {
            if (!deleteCtx) return;
            const match = this.value === deleteCtx.name;
            document.getElementById('dbi-delete-error').classList.toggle('d-none', !this.value || match);
            document.getElementById('dbi-delete-confirm').disabled = !match;
        });
        document.getElementById('dbi-delete-confirm').addEventListener('click', deleteInstance);
        return refs;
    }

    function resetSelect(sel, placeholder) {
        sel.replaceChildren();
        const opt = document.createElement('option');
        opt.value = ''; opt.textContent = placeholder || '-';
        sel.appendChild(opt);
    }

    // ── 생성 폼 ───────────────────────────────────────────
    function openCreate(panel) {
        const r = uiInit();
        if (!r) return;
        active = panel;
        engineVersionData = [];
        const provider = panel.getProvider();
        document.getElementById('dbi-context').textContent =
            (panel.getCredLabel && panel.getCredLabel()) || `${provider} / ${panel.getRegion()}`;
        // 폼 초기화 + provider별 규칙
        resetSelect(r.engine); resetSelect(r.version); icReset();
        r.instanceId.value = '';
        r.instanceId.classList.remove('is-invalid');
        r.instanceId.closest('.form-floating').classList.remove('is-invalid');
        r.username.value = '';
        r.password.value = '';
        ['dbi-username-error', 'dbi-password-error'].forEach(id => {
            const el = document.getElementById(id);
            el.textContent = ''; el.classList.add('d-none');
        });
        const p = (provider || '').toLowerCase();
        // NoSQL aws(DynamoDB)/gcp(Firestore)는 서버리스 — Instance ID만으로 생성 (gen-no-sql과 동일)
        const simple = isSimpleCreate(panel, p);
        // engine 셀렉트는 rdbms만 (nrdbms는 engineVersion만 존재)
        document.getElementById('dbi-engine-group').classList.toggle('d-none', simple || panel.kind !== 'rdbms');
        document.getElementById('dbi-version-group').classList.toggle('d-none', simple);
        document.getElementById('dbi-class-group').classList.toggle('d-none', simple);
        document.getElementById('dbi-username-group').classList.toggle('d-none', simple);
        document.getElementById('dbi-password-group').classList.toggle('d-none', simple);
        // storage: rdbms gcp=10GB~, 그 외 20GB~. nrdbms ncp는 불필요 → 숨김
        const storageHidden = simple || (panel.kind === 'nrdbms' && p === 'ncp');
        document.getElementById('dbi-storage-group').classList.toggle('d-none', storageHidden);
        r.storage.min = (panel.kind === 'rdbms' && p === 'gcp') ? '10' : '20';
        r.storage.value = r.storage.min;
        document.getElementById('dbi-storage-error').textContent = `Please enter ${r.storage.min} GB or more.`;
        r.storage.classList.remove('is-invalid');
        r.storage.closest('.form-floating').classList.remove('is-invalid');
        // username: gen-mysql과 동일 — GCP도 강제 채움/readonly 없이 검증+안내로만 처리
        const help = simple ? '' : usernameHelp(panel.kind, provider);
        document.getElementById('dbi-username-help').textContent = help;
        // 여백: 헬퍼 문구가 바로 아래 붙는 provider는 mb-1, 그 외 mb-3 (gen과 동일)
        const unGrp = document.getElementById('dbi-username-group');
        unGrp.classList.toggle('mb-1', !!help);
        unGrp.classList.toggle('mb-3', !help);
        updateCreateBtn();
        r.oc.show();
        if (!simple) loadEngineVersions();
    }

    // NoSQL aws/gcp: 프로비저닝 스펙 없이 Instance ID만으로 생성 (백엔드도 instanceId만 필수)
    function isSimpleCreate(panel, providerLower) {
        return panel.kind === 'nrdbms' && (providerLower === 'aws' || providerLower === 'gcp');
    }

    // 기존 인스턴스와 중복된 Instance ID인지 (페이지가 목록을 getRows로 제공)
    function instanceIdDup() {
        const v = refs.instanceId.value.trim();
        if (!v || !active?.getRows) return false;
        return (active.getRows() || []).some(i => i.instanceId === v || i.name === v);
    }

    async function loadEngineVersions() {
        if (!active) return;
        try {
            const res = await fetch(`/db/${active.kind}/engine-versions`, {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ provider: active.getProvider(), region: active.getRegion() })
            });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            engineVersionData = await res.json() || [];
            if (active.kind === 'rdbms') {
                const engines = [...new Set(engineVersionData.map(d => d.engine))];
                resetSelect(refs.engine);
                engines.forEach(e => {
                    const opt = document.createElement('option');
                    opt.value = e; opt.textContent = e;
                    refs.engine.appendChild(opt);
                });
                if (engines.length) { refs.engine.selectedIndex = 1; }
                populateVersions();
            } else {
                resetSelect(refs.version);
                engineVersionData.forEach(d => {
                    const opt = document.createElement('option');
                    opt.value = d.engineVersion; opt.textContent = d.engineVersion;
                    refs.version.appendChild(opt);
                });
                if (engineVersionData.length) {
                    refs.version.selectedIndex = 1;
                    loadClasses();
                }
            }
        } catch (e) { console.error('InstancePanel.engineVersions:', e); }
    }

    function populateVersions() {
        const versions = engineVersionData
            .filter(d => d.engine === refs.engine.value)
            .map(d => d.engineVersion);
        resetSelect(refs.version);
        versions.forEach(v => {
            const opt = document.createElement('option');
            opt.value = v; opt.textContent = v;
            refs.version.appendChild(opt);
        });
        icReset();
        if (versions.length) {
            refs.version.selectedIndex = 1;
            loadClasses();
        }
    }

    async function loadClasses() {
        if (!active || !refs.version.value) { icReset(); return; }
        const provider = active.getProvider();
        const body = { provider, region: active.getRegion(), engineVersion: refs.version.value };
        if (active.kind === 'rdbms') body.engine = refs.engine.value;
        try {
            const res = await fetch(`/db/${active.kind}/instance-class`, {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            icLoadFamilies(await res.json() || []);
        } catch (e) { console.error('InstancePanel.instanceClass:', e); }
    }

    // ── Instance Class 2패널 드롭다운 (gen-mysql/gen-no-sql과 동일 UX) ──
    function icClose() { refs.icPanel.classList.add('d-none'); }

    function icReset() {
        icClasses = [];
        refs.klass.value = '';
        refs.icTriggerText.textContent = '-';
        refs.icTriggerText.classList.add('text-muted');
        refs.icFamilyList.innerHTML = '';
        refs.icSizeList.innerHTML = '';
        icClose();
    }

    function icSelectFamily(family, familyEl) {
        const provider = active ? active.getProvider() : '';
        refs.icFamilyList.querySelectorAll('.ic-family-item').forEach(el => el.classList.remove('active'));
        familyEl.classList.add('active');
        const sep = (provider || '').toLowerCase() === 'gcp' ? '-' : '.';
        const sizes = icClasses.filter(c => c.startsWith(family + sep));
        refs.icSizeList.innerHTML = '';
        sizes.forEach(s => {
            const div = document.createElement('div');
            div.className = 'ic-size-item px-3 py-2 small';
            div.textContent = icSplit(s, provider).size;
            div.addEventListener('click', e => {
                e.stopPropagation();
                refs.klass.value = s;
                refs.icTriggerText.textContent = s;
                refs.icTriggerText.classList.remove('text-muted');
                icClose();
                updateCreateBtn();
            });
            refs.icSizeList.appendChild(div);
        });
    }

    function icLoadFamilies(classes) {
        const provider = active ? active.getProvider() : '';
        icReset();
        icClasses = Array.isArray(classes) ? classes : [];
        const families = [...new Set(icClasses.map(c => icSplit(c, provider).family))];
        families.forEach(f => {
            const div = document.createElement('div');
            div.className = 'ic-family-item px-3 py-2 small';
            div.textContent = f;
            div.addEventListener('click', e => { e.stopPropagation(); icSelectFamily(f, div); });
            refs.icFamilyList.appendChild(div);
        });
        // gen과 동일: 첫 family/size 자동 선택
        const firstEl = refs.icFamilyList.querySelector('.ic-family-item');
        if (firstEl) {
            icSelectFamily(families[0], firstEl);
            const firstSize = refs.icSizeList.querySelector('.ic-size-item');
            if (firstSize) firstSize.click();
        }
        updateCreateBtn();
    }

    function updateCreateBtn() {
        if (!active || !refs) return;
        const provider = active.getProvider();
        // NoSQL aws/gcp: Instance ID만으로 생성 가능
        if (isSimpleCreate(active, (provider || '').toLowerCase())) {
            refs.createBtn.disabled = !refs.instanceId.value.trim() || instanceIdDup();
            return;
        }
        const storageHidden = document.getElementById('dbi-storage-group').classList.contains('d-none');
        const min = Number(refs.storage.min) || 20;
        const storageOk = storageHidden || (refs.storage.value !== '' && Number(refs.storage.value) >= min);
        const un = refs.username.value.trim();
        const pw = refs.password.value;
        const ok = refs.instanceId.value.trim()
            && !instanceIdDup()
            && refs.version.value
            && refs.klass.value
            && storageOk
            && un && !usernameError(active.kind, provider, un)
            && pw && !passwordError(provider, pw);
        refs.createBtn.disabled = !ok;
    }

    async function createInstance() {
        if (!active) return;
        const panel = active;
        const body = {
            provider: panel.getProvider(),
            region: panel.getRegion(),
            instanceId: refs.instanceId.value.trim(),
            instanceClass: refs.klass.value,
            engineVersion: refs.version.value,
            masterUsername: refs.username.value.trim(),
            masterPassword: refs.password.value,
            allocatedStorage: parseInt(refs.storage.value, 10) || 0,
        };
        if (panel.kind === 'rdbms') body.engine = refs.engine.value;
        const btn = refs.createBtn;
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>&nbsp;Creating...';
        try {
            const res = await fetch(`/db/${panel.kind}`, {
                method: 'PUT', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });
            const json = await res.json().catch(() => ({}));
            if (!res.ok) throw new Error(json.error || `HTTP ${res.status}`);
            refs.oc.hide();
            const settle = CREATE_SETTLE_MS[(panel.getProvider() || '').toLowerCase()] || 0;
            startPoll(panel, 'create', json.instanceId || body.instanceId, settle);
        } catch (err) {
            alert('Failed to create instance: ' + err.message);
        } finally {
            btn.disabled = false;
            btn.textContent = 'Create';
        }
    }

    // ── 삭제 ─────────────────────────────────────────────
    function openDelete(panel) {
        const r = uiInit();
        if (!r) return;
        const inst = panel.getSelected && panel.getSelected();
        if (!inst) return;
        deleteCtx = { panel, instanceId: inst.instanceId, name: inst.name || inst.instanceId };
        document.getElementById('dbi-delete-name').textContent = inst.instanceId;
        document.getElementById('dbi-delete-instruction').textContent =
            'To delete this instance, enter its Name: ' + deleteCtx.name;
        r.deleteModal.show();
    }

    async function deleteInstance() {
        if (!deleteCtx) return;
        const { panel, instanceId } = deleteCtx;
        const btn = document.getElementById('dbi-delete-confirm');
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>Deleting...';
        try {
            const res = await fetch(`/db/${panel.kind}`, {
                method: 'DELETE', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ provider: panel.getProvider(), region: panel.getRegion(), instanceId })
            });
            const json = await res.json().catch(() => ({}));
            if (!res.ok) throw new Error(json.error || 'Delete failed');
            refs.deleteModal.hide();
            startPoll(panel, 'delete', instanceId);
        } catch (err) {
            alert('Failed to delete instance: ' + err.message);
        } finally {
            btn.disabled = false;
            btn.innerHTML = '<i class="bi bi-trash3 me-1"></i>Delete';
        }
    }

    // ── 완료 폴링: 목록을 재조회해 onData로 넘기고, 전이가 끝나면 중단 ──
    // initialDelayMs: 첫 틱만 지연시킨다 (0/생략이면 기존대로 즉시 실행)
    function startPoll(panel, mode, instanceId, initialDelayMs) {
        stopPoll();
        const provider = panel.getProvider(), region = panel.getRegion();
        let ticks = 0;
        const MAX_TICKS = 240; // 5s × 240 = 20분 상한 (프로비저닝 지연/유실 대비)
        async function tick() {
            // 사용자가 다른 credential/region으로 이동했으면 중단
            if (panel.getProvider() !== provider || panel.getRegion() !== region) { stopPoll(); return; }
            if (++ticks > MAX_TICKS) { stopPoll(); return; }
            try {
                const res = await fetch(`/db/${panel.kind}`, {
                    method: 'POST', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ provider, region })
                });
                if (res.ok) {
                    const instances = await res.json() || [];
                    panel.onData(instances);
                    panel.refreshButtons(); // 상태 전이(creating→available 등)에 따라 Delete 버튼 재평가
                    const found = Array.isArray(instances) ? instances.find(i => i.instanceId === instanceId) : null;
                    const done = mode === 'delete'
                        ? !found
                        : (found && ['available', 'failed'].includes((found.status || '').toLowerCase()));
                    if (done) { stopPoll(); return; }
                }
            } catch (e) { console.error('InstancePanel.poll:', e); }
            pollTimer = setTimeout(tick, POLL_MS);
        }
        if (initialDelayMs > 0) pollTimer = setTimeout(tick, initialDelayMs);
        else tick();
    }
    function stopPoll() {
        if (pollTimer) { clearTimeout(pollTimer); pollTimer = null; }
    }

    // ── init ─────────────────────────────────────────────
    function init(cfg) {
        const panel = {
            kind: cfg.kind,
            getProvider: cfg.getProvider,
            getRegion: cfg.getRegion,
            getCredLabel: cfg.getCredLabel,
            getSelected: cfg.getSelected,
            getRows: cfg.getRows,       // 중복 Instance ID 검증용 현재 목록 (옵션)
            onData: cfg.onData,
            createBtn: cfg.createBtn || null,
            deleteBtn: cfg.deleteBtn || null,
        };
        // 샘플데이터 생성 연동 (cfg.sample 지정 시, SamplePanel 모듈):
        // 인스턴스 선택 시 '+ Instance' 버튼이 '+ Sample Data'로 전환된다 (Generate 페이지와 동일 UX).
        // cfg.sample = {} (SQL) 또는 { awsBtn: el } (NoSQL — AWS 서버리스 진입 버튼)
        const sampleCfg = cfg.sample ? {
            kind: cfg.kind,
            getProvider: cfg.getProvider, getRegion: cfg.getRegion,
            getCredLabel: cfg.getCredLabel, getSelected: cfg.getSelected,
        } : null;
        // 버튼 전환은 '선택된 인스턴스'가 있을 때만 (NoSQL AWS는 목록이 없어 전용 버튼으로 진입)
        const sampleEligible = () => !!sampleCfg
            && !!(cfg.getSelected && cfg.getSelected())
            && SamplePanel.canOpen(sampleCfg);
        panel.refreshButtons = () => {
            const p = panel.getProvider(), r = panel.getRegion();
            const ctxOk = p && p !== 'none' && r && r !== 'none';
            if (panel.createBtn) panel.createBtn.disabled = !ctxOk;
            // 생성 중(creating/net_creating 등) 인스턴스는 삭제 불가 — available 전이 후 재활성
            const sel = panel.getSelected && panel.getSelected();
            const row = sel && panel.getRows
                ? (panel.getRows() || []).find(i => i.instanceId === sel.instanceId) : null;
            const isCreating = !!row && (row.status || '').toLowerCase().includes('creating');
            if (panel.deleteBtn) panel.deleteBtn.disabled = !(ctxOk && sel) || isCreating;
            if (panel.createBtn && sampleCfg) {
                const s = sampleEligible();
                panel.createBtn.innerHTML = '<i class="bi bi-plus-lg me-1"></i>' + (s ? 'Sample Data' : 'Instance');
                if (s) panel.createBtn.disabled = false;
            }
            if (sampleCfg && cfg.sample.awsBtn) cfg.sample.awsBtn.disabled = !SamplePanel.canOpen(sampleCfg);
        };
        if (panel.createBtn) panel.createBtn.addEventListener('click', () => {
            if (sampleEligible()) SamplePanel.open(sampleCfg);
            else openCreate(panel);
        });
        if (sampleCfg && cfg.sample.awsBtn) cfg.sample.awsBtn.addEventListener('click', () => SamplePanel.open(sampleCfg));
        if (panel.deleteBtn) panel.deleteBtn.addEventListener('click', () => openDelete(panel));
        panel.refreshButtons();
        return panel;
    }

    return { init };
})();


// ─────────────────────────────────────────────────────────────────────────────
// SamplePanel — 샘플데이터 생성 공용 컴포넌트 (SQL·NoSQL)
//
// gen-mysql/gen-no-sql의 Sample Data 생성 기능(제출 API·provider별 필드·검증)을
// migration/backup/restore 페이지에서 재사용하기 위한 모듈.
// InstancePanel과 연동해 Generate 페이지와 동일한 UX를 제공한다:
// 인스턴스 선택 시 '+ Instance' 버튼이 '+ Sample Data'로 전환되어 이 패널을 연다.
//
// 사용법 1 (권장 — InstancePanel 연동):
//   InstancePanel.init({ kind, ..., sample: {} })                 // SQL
//   InstancePanel.init({ kind, ..., sample: { awsBtn: el } })     // NoSQL (AWS 서버리스 진입 버튼)
// 사용법 2 (독립 버튼):
//   SamplePanel.attach({ openBtn, kind, getProvider, getRegion, getCredLabel?, getSelected })
//
// provider별 규칙 (Generate 페이지와 동일):
//   [SQL/rdbms]   선택한 인스턴스의 Host 준비 후(alibaba/ncp는 공인 'pub') 계정 입력 → /generate/rdbms
//   [NoSQL/nrdbms] AWS: 서버리스 — Instance ID만 입력
//                  GCP: Database ID = 선택한 Instance ID (읽기 전용)
//                  NCP/Alibaba: 공인('pub') 호스트 준비 후 계정 + Database Name 입력
//                  → /generate/nrdbms
// ─────────────────────────────────────────────────────────────────────────────
const SamplePanel = (() => {
    let refs = null;
    let active = null; // 패널을 연 cfg { kind, getProvider, getRegion, getCredLabel?, getSelected }

    // provider별 username 규칙 (Generate 페이지와 동일 기준)
    function usernameHelp(kind, p) {
        if (kind === 'rdbms') return p === 'gcp' ? 'Only "root" is allowed.' : '';
        if (p === 'alibaba') return 'Only "root" is allowed.';
        if (p === 'ncp') return '"admin" and "root" are not allowed.';
        return '';
    }
    function usernameError(kind, p, val) {
        const v = (val || '').trim();
        if (!v) return '';
        const lc = v.toLowerCase();
        if (kind === 'rdbms') {
            if (p === 'gcp' && lc !== 'root') return 'Only "root" is allowed for GCP.';
            return '';
        }
        if (p === 'alibaba' && lc !== 'root') return 'Only "root" is allowed for Alibaba.';
        if (p === 'ncp' && (lc === 'admin' || lc === 'root')) return '"admin" and "root" are not allowed for NCP.';
        return '';
    }
    // Host가 샘플데이터 접속에 사용 가능한지 (alibaba/ncp: 공인 'pub' 도메인 필요)
    function hostReady(endpoint, p) {
        const h = endpoint || '';
        if (p === 'alibaba' || p === 'ncp') return h.includes('pub');
        return !!h;
    }
    function providerOf(cfg) { return ((cfg || active)?.getProvider() || '').toLowerCase(); }
    function kindOf(cfg) { return (cfg || active)?.kind || 'nrdbms'; }
    function isMongo(cfg) {
        const p = providerOf(cfg);
        return kindOf(cfg) === 'nrdbms' && (p === 'ncp' || p === 'alibaba');
    }
    // 계정 입력(username/password)이 필요한 조합: SQL 전체 + NoSQL MongoDB
    function needsAccount(cfg) { return kindOf(cfg) === 'rdbms' || isMongo(cfg); }

    function uiInit() {
        if (refs) return refs;
        const oc = document.getElementById('nspOffcanvas');
        if (!oc) { console.warn('SamplePanel: sample-data-panel.html 파셜이 include되지 않았습니다'); return null; }
        refs = {
            oc: bootstrap.Offcanvas.getOrCreateInstance(oc),
            context:       document.getElementById('nsp-context'),
            gcpGroup:      document.getElementById('nsp-gcp-group'),
            awsGroup:      document.getElementById('nsp-aws-group'),
            accountFields: document.getElementById('nsp-account-fields'),
            dbGroup:       document.getElementById('nsp-db-group'),
            databaseId:    document.getElementById('nsp-databaseId'),
            instanceId:    document.getElementById('nsp-instanceId'),
            username:      document.getElementById('nsp-username'),
            usernameHelp:  document.getElementById('nsp-username-help'),
            usernameErr:   document.getElementById('nsp-username-error'),
            password:      document.getElementById('nsp-password'),
            database:      document.getElementById('nsp-database'),
            submitBtn:     document.getElementById('nsp-submitBtn'),
        };
        refs.username.addEventListener('input', () => {
            const msg = usernameError(kindOf(), providerOf(), refs.username.value);
            refs.usernameErr.textContent = msg;
            refs.usernameErr.classList.toggle('d-none', !msg);
            updateSubmit();
        });
        refs.password.addEventListener('input', updateSubmit);
        refs.database.addEventListener('input', updateSubmit);
        refs.instanceId.addEventListener('input', updateSubmit);
        refs.submitBtn.addEventListener('click', submit);
        return refs;
    }

    // Sample Data 진입 가능 조건 (Generate 페이지의 행 선택 게이트와 동일 기준)
    function canOpen(cfg) {
        const p = providerOf(cfg);
        const r = cfg.getRegion() || '';
        if (!p || p === 'none' || !r || r === 'none') return false;
        const sel = cfg.getSelected?.() || null;
        if (kindOf(cfg) === 'rdbms') {
            // SQL: 선택한 인스턴스의 Host 준비 필요 (alibaba/ncp는 공인 'pub')
            return !!sel && hostReady(sel.endpoint, p);
        }
        if (p === 'aws') return true;               // 서버리스 — 인스턴스 선택 불필요
        if (!sel) return false;
        if (p === 'gcp') return true;               // Database ID = Instance ID
        return hostReady(sel.endpoint, p);          // ncp/alibaba: 공인 도메인 준비 후
    }

    function open(cfg) {
        if (!uiInit()) return;
        active = cfg;
        const kind = kindOf();
        const p = providerOf();
        const sel = cfg.getSelected?.() || null;
        const cred = cfg.getCredLabel?.() || p;
        refs.context.textContent = sel ? `${cred} / ${sel.name || sel.instanceId}` : cred;
        // kind/provider별 필드 토글
        refs.gcpGroup.classList.toggle('d-none', !(kind === 'nrdbms' && p === 'gcp'));
        refs.awsGroup.classList.toggle('d-none', !(kind === 'nrdbms' && p === 'aws'));
        refs.accountFields.classList.toggle('d-none', !needsAccount());
        refs.dbGroup.classList.toggle('d-none', !isMongo());
        // 입력 초기화 (열 때마다 빈 폼 — 다른 인스턴스 값 잔존 방지)
        refs.databaseId.value = (kind === 'nrdbms' && p === 'gcp') ? ((sel && sel.instanceId) || '') : '';
        refs.instanceId.value = '';
        refs.username.value = '';
        refs.usernameErr.classList.add('d-none');
        refs.usernameHelp.textContent = usernameHelp(kind, p);
        refs.password.value = '';
        refs.database.value = '';
        updateSubmit();
        refs.oc.show();
    }

    function updateSubmit() {
        if (!refs || !active) return;
        const kind = kindOf();
        const p = providerOf();
        const un = refs.username.value.trim();
        const pw = refs.password.value.trim();
        let ok;
        if (kind === 'rdbms') ok = !!un && !!pw; // gen-mysql phase 3와 동일 게이트 (세부 검증은 API 응답)
        else if (p === 'gcp') ok = !!refs.databaseId.value.trim();
        else if (p === 'aws') ok = !!refs.instanceId.value.trim();
        else ok = !!un && !usernameError(kind, p, un) && !!pw && !!refs.database.value.trim();
        refs.submitBtn.disabled = !ok;
    }

    // 샘플데이터 생성 — gen-mysql(/generate/rdbms)·gen-no-sql(/generate/nrdbms)과 동일 payload
    async function submit() {
        const kind = kindOf();
        const p = providerOf();
        const sel = active?.getSelected?.() || null;
        const targetPoint = {
            provider: p,
            region:   active.getRegion() || '',
            host:     needsAccount() ? (sel?.endpoint || '') : '',
            port:     String(needsAccount() ? (sel?.port || '') : ''),
            username: refs.username.value.trim(),
            password: refs.password.value.trim(),
            instanceId: (kind === 'nrdbms' && p === 'aws')
                ? refs.instanceId.value.trim()
                : ((sel && sel.instanceId) || ''),
        };
        if (kind === 'nrdbms') {
            targetPoint.databaseName  = refs.database.value.trim();
            targetPoint.databaseId    = refs.databaseId.value.trim();
            targetPoint.engineVersion = '';
            targetPoint.instanceClass = '';
        }
        // 생성 중 중복 요청 방지: 비활성화 + 스피너 + In progress (Generate 페이지와 동일 UX)
        refs.submitBtn.disabled = true;
        refs.submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>&nbsp;In progress..';
        try {
            const res = await fetch(kind === 'rdbms' ? '/generate/rdbms' : '/generate/nrdbms', {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ targetPoint, dummy: targetPoint })
            });
            if (!res.ok) {
                const errJson = await res.json().catch(() => ({}));
                throw new Error(errJson.Result || errJson.error || `HTTP ${res.status}`);
            }
            const json = await res.json().catch(() => ({}));
            // 결과는 페이지 공용 Result 패널(#resultText, result.html 파셜)에 표시 — Generate 페이지와 동일 UX
            const rt = document.getElementById('resultText');
            if (rt) rt.value = json.Result || 'Sample data generated.';
            resultCollpase();
            refs.oc.hide(); // 성공 → offcanvas를 닫아 Result 패널 노출
        } catch (err) {
            const rt = document.getElementById('resultText');
            if (rt) rt.value = err.message || String(err);
            resultCollpase();
            alert('Failed to generate sample data. See the Result panel for details.');
        } finally {
            // 결과 표시 후 스피너 제거 → 원래 라벨 복원 → 폼 기준 재활성화
            refs.submitBtn.textContent = 'Submit';
            updateSubmit();
        }
    }

    // 독립 버튼 진입점 (예: NoSQL AWS 안내 배너의 Sample Data 버튼)
    function attach(cfg) {
        cfg.openBtn?.addEventListener('click', () => open(cfg));
        const api = {
            refreshButtons: () => { if (cfg.openBtn) cfg.openBtn.disabled = !canOpen(cfg); },
        };
        api.refreshButtons();
        return api;
    }

    return { attach, canOpen, open };
})();
