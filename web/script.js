function escapeHtml(unsafe) {
  unsafe = unsafe.toString();
  if (typeof unsafe !== "string") {
    throw new Error("Cannot escape " + typeof unsafe);
  }
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// handle changing the filename in the input, reflect the changes in the examples

function updateFilename(val, target = null) {
  localStorage.setItem("filename", val);
  document.querySelector(".filename-input").value = val;
  Array.from(document.querySelectorAll(".filename-input")).forEach((elem) => {
    if (elem !== target) {
      elem.value = val;
    }
  });
  Array.from(document.querySelectorAll(".filename-placeholder")).forEach(
    (elem) => (elem.innerHTML = escapeHtml(val))
  );
}
Array.from(document.querySelectorAll(".filename-input")).forEach((i) =>
  i.addEventListener("keyup", (ev) => {
    updateFilename(ev.target.value, ev.target);
  })
);
if (localStorage.getItem("filename")) {
  updateFilename(localStorage.getItem("filename"));
}

// handle drop area

let dropArea = document.querySelector(".drop-upload");

["dragenter", "dragover", "dragleave", "drop"].forEach((eventName) => {
  dropArea.addEventListener(eventName, preventDefaults, false);
});

function preventDefaults(e) {
  e.preventDefault();
  e.stopPropagation();
}

["dragenter", "dragover"].forEach((eventName) => {
  dropArea.addEventListener(eventName, highlight, false);
});
["dragleave", "drop"].forEach((eventName) => {
  dropArea.addEventListener(eventName, unhighlight, false);
});

function highlight(e) {
  dropArea.classList.add("highlight");
}

function unhighlight(e) {
  dropArea.classList.remove("highlight");
}

dropArea.addEventListener("drop", handleDrop, false);

let fileToUpload = null;

function handleDrop(e) {
  let dt = e.dataTransfer;
  let files = dt.files;

  updateFilename(files[0].name);
  fileToUpload = files[0];
  document.querySelector(".upload-confirm").disabled = false;
  doUpload();
}

// handle the Browse... input

document.querySelector(".upload-input").addEventListener("change", (ev) => {
  let files = ev.target.files;
  updateFilename(files[0].name);
  fileToUpload = files[0];
  document.querySelector(".upload-confirm").disabled = false;
  doUpload();
});

function doUpload() {
  var request = new XMLHttpRequest();
  let filename = document.querySelector(".filename-input").value;
  if (filename[0] !== "/") {
    filename = "/" + filename;
  }
  request.open("PUT", filename);
  request.send(fileToUpload);
  request.upload.addEventListener(
    "progress",
    (event) => {
      var percent = Math.round((event.loaded / event.total) * 100);
      console.log({ percent });
      document.querySelector(".upload-progress").style.width = percent + "%";
    },
    false
  );
  request.addEventListener(
    "load",
    () => {
      document.querySelector(".upload-progress").style.width = "100%";
      setTimeout(() => {
        document.querySelector(".upload-progress").style.width = "0%";
      }, 1000);
    },
    false
  );
  fetchRecents();
}

// handle recents

async function fetchRecents(wait = false) {
  const resp = await fetch("/recents.json" + (wait ? "?wait" : ""));
  if (!resp.ok) {
    throw new Error("Server returned an error status!");
  }
  const recents = await resp.json();
  console.log(recents);
  document.querySelector(".recent-uploads").innerHTML = recents
    .map((r) => ({ ...r, uploadedAt: new Date(r.uploadedAt) }))
    .map(
      (r) => `
    <tr>
    <td>${r.uploadedAt.getFullYear()}-${(r.uploadedAt.getMonth() + 1)
        .toString()
        .padStart(2, "0")}-${r.uploadedAt
        .getDay()
        .toString()
        .padStart(2, "0")} ${r.uploadedAt
        .getHours()
        .toString()
        .padStart(2, "0")}:${r.uploadedAt
        .getMinutes()
        .toString()
        .padStart(2, "0")}</td>  
    <td><a href="${escapeHtml(r.url)}">${escapeHtml(r.name)}</td>
    <td title="${escapeHtml(r.sizeExact)} bytes">${escapeHtml(r.size)}</td>
    <td><div class="trunc"><a href="${escapeHtml(
      r.url + "?info"
    )}" title="${escapeHtml(r.type)}">${escapeHtml(r.type)}</div></td>
    <td>${escapeHtml(r.uploaderLocation)}</td>
    
      </tr>
    `
    )
    .join("\n");
}
async function run() {
  await fetchRecents();
  while (true) {
    await new Promise((res) => setTimeout(res, 2000));
    await fetchRecents(true);
  }
}
run().catch(console.error);

// handle http/https switching

function updateHttpsHttpState() {
  const isHttp = document.querySelector(".url-proto").textContent === "http://";
  let a = document.querySelector(isHttp ? ".btn-http" : ".btn-https");
  let b = document.querySelector(!isHttp ? ".btn-http" : ".btn-https");
  a.classList.add("active");
  b.classList.remove("active");
  document.querySelector(".btn-tftp").classList.remove("active");
  hide(".tftp");
  show(".http-or-https");
  document.querySelector(".bttns-curl-wget").style.display = "flex";
}

document.querySelector(".btn-http").addEventListener("click", () => {
  Array.from(document.querySelectorAll(".url-proto")).forEach(
    (elem) => (elem.textContent = "http://")
  );
  updateHttpsHttpState();
});

document.querySelector(".btn-https").addEventListener("click", () => {
  Array.from(document.querySelectorAll(".url-proto")).forEach(
    (elem) => (elem.textContent = "https://")
  );
  updateHttpsHttpState();
});

function show(selector) {
  Array.from(document.querySelectorAll(selector)).forEach(
    (elem) => (elem.style.display = "block")
  );
}

function hide(selector) {
  Array.from(document.querySelectorAll(selector)).forEach(
    (elem) => (elem.style.display = "none")
  );
}

document.querySelector(".btn-wget").addEventListener("click", function () {
  show(".http-or-https");
  show(".wget");
  hide(".curl");
  hide(".tftp");
  this.classList.add("active");
  document.querySelector(".btn-curl").classList.remove("active");
  document.querySelector(".btn-tftp").classList.remove("active");
});

document.querySelector(".btn-curl").addEventListener("click", function () {
  show(".http-or-https");
  show(".curl");
  hide(".wget");
  hide(".tftp");
  this.classList.add("active");
  document.querySelector(".btn-wget").classList.remove("active");
  document.querySelector(".btn-tftp").classList.remove("active");
});

document.querySelector(".btn-tftp").addEventListener("click", function () {
  hide(".http-or-https");
  show(".tftp");
  hide(".bttns-curl-wget");
  this.classList.add("active");
  document.querySelector(".btn-wget").classList.remove("active");
  document.querySelector(".btn-curl").classList.remove("active");
  document.querySelector(".btn-http").classList.remove("active");
  document.querySelector(".btn-https").classList.remove("active");
});

updateHttpsHttpState();
