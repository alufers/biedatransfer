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
function updateFilename(val) {
  localStorage.setItem("filename", val);
  document.querySelector(".filename-input").value = val;
  Array.from(document.querySelectorAll(".filename-placeholder")).forEach(
    (elem) => (elem.innerHTML = escapeHtml(val))
  );
}
document.querySelector(".filename-input").addEventListener("keyup", (ev) => {
  updateFilename(ev.target.value);
});
if (localStorage.getItem("filename")) {
  updateFilename(localStorage.getItem("filename"));
}

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
      document.querySelector(".upload-progress").style.width =
        percent + "percent";
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
