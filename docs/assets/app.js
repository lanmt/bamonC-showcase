const files = [
  "main.go",
  "routers/initrouter.go",
  "task/corn.go",
  "task/captcha_task.go",
  "request/helper.go",
  "request/submit_reservation.go",
  "service/user_service.go",
  "model/user.go"
];

const list = document.querySelector("#fileList");
const output = document.querySelector("#sourceCode");
const current = document.querySelector("#currentFile");
const copy = document.querySelector("#copyCode");

function escapeHTML(value) {
  return value.replace(/[&<>]/g, character => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;" })[character]);
}

function highlight(source) {
  return escapeHTML(source)
    .replace(/(".*?")/g, '<span class="token-string">$1</span>')
    .replace(/\b(package|import|func|var|const|type|struct|interface|return|if|else|for|range|go|defer|select|case|switch)\b/g, '<span class="token-keyword">$1</span>')
    .replace(/(\/\/[^\n]*)/g, '<span class="token-comment">$1</span>');
}

async function showFile(file, button) {
  document.querySelectorAll(".file-list button").forEach(item => item.classList.remove("active"));
  button.classList.add("active");
  current.textContent = file;
  output.textContent = "正在加载…";
  try {
    const response = await fetch(`source/${file}.txt`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    output.innerHTML = highlight(await response.text());
  } catch (error) {
    output.textContent = `无法加载源码：${error.message}`;
  }
}

files.forEach((file, index) => {
  const button = document.createElement("button");
  button.type = "button";
  button.role = "tab";
  button.textContent = file;
  button.addEventListener("click", () => showFile(file, button));
  list.appendChild(button);
  if (index === 0) showFile(file, button);
});

copy.addEventListener("click", async () => {
  await navigator.clipboard.writeText(output.textContent);
  copy.textContent = "已复制";
  setTimeout(() => { copy.textContent = "复制代码"; }, 1200);
});
