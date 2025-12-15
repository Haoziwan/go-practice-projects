let progressInterval = null;

// Update workers value display
document.getElementById("workers").addEventListener("input", (e) => {
  document.getElementById("workers-value").textContent = e.target.value;
});

// Auto-fill filename from URL
document.getElementById("url").addEventListener("input", (e) => {
  const url = e.target.value.trim();
  if (url) {
    try {
      const urlObj = new URL(url);
      const pathname = urlObj.pathname;
      const filename = pathname.substring(pathname.lastIndexOf("/") + 1);
      if (filename && !document.getElementById("filename").value) {
        document.getElementById("filename").value = filename;
      }
    } catch (err) {
      // Invalid URL, ignore
    }
  }
});

// Start download
document.getElementById("download-btn").addEventListener("click", async () => {
  const url = document.getElementById("url").value.trim();
  const filename = document.getElementById("filename").value.trim();
  const workers = parseInt(document.getElementById("workers").value);

  if (!url) {
    alert("请输入下载链接");
    return;
  }

  if (!filename) {
    alert("请输入文件名");
    return;
  }

  try {
    // Disable button
    const btn = document.getElementById("download-btn");
    btn.disabled = true;
    btn.innerHTML = '<span class="btn-icon">⏳</span> 正在启动...';

    // Start download
    await window.go.main.App.StartDownload(url, filename, workers);

    // Show progress card
    document.getElementById("progress-card").style.display = "block";

    // Start progress updates
    startProgressUpdates();

    // Re-enable button after a delay
    setTimeout(() => {
      btn.disabled = false;
      btn.innerHTML = '<span class="btn-icon">⬇</span> 开始下载';
    }, 2000);
  } catch (error) {
    alert("启动下载失败: " + error);
    const btn = document.getElementById("download-btn");
    btn.disabled = false;
    btn.innerHTML = '<span class="btn-icon">⬇</span> 开始下载';
  }
});

// Cancel download
document.getElementById("cancel-btn").addEventListener("click", async () => {
  try {
    await window.go.main.App.CancelDownload();
    stopProgressUpdates();
    document.getElementById("progress-card").style.display = "none";
    resetProgress();
  } catch (error) {
    alert("取消下载失败: " + error);
  }
});

function startProgressUpdates() {
  if (progressInterval) {
    clearInterval(progressInterval);
  }

  progressInterval = setInterval(async () => {
    try {
      const progress = await window.go.main.App.GetProgress();
      updateProgressUI(progress);

      if (progress.status === "completed" || progress.status === "error") {
        stopProgressUpdates();
        if (progress.status === "completed") {
          setTimeout(() => {
            document.getElementById("progress-card").style.display = "none";
            resetProgress();
          }, 1500);
        }
      }
    } catch (error) {
      console.error("Failed to get progress:", error);
    }
  }, 500);
}

function stopProgressUpdates() {
  if (progressInterval) {
    clearInterval(progressInterval);
    progressInterval = null;
  }
}

function updateProgressUI(progress) {
  // Update progress bar
  const progressBar = document.getElementById("progress-bar");
  const progressText = document.getElementById("progress-text");
  progressBar.style.width = progress.percentage.toFixed(1) + "%";
  progressText.textContent = progress.percentage.toFixed(1) + "%";

  // Update status
  const statusEl = document.getElementById("status");
  statusEl.textContent = getStatusText(progress.status);
  statusEl.className = "status " + progress.status;

  // Update stats
  document.getElementById("speed").textContent =
    progress.speed.toFixed(2) + " MB/s";
  document.getElementById("downloaded").textContent =
    progress.downloaded.toFixed(2) + " MB";
  document.getElementById("total").textContent =
    progress.total.toFixed(2) + " MB";

  // Update remaining time
  if (progress.remainingTime > 0) {
    const minutes = Math.floor(progress.remainingTime / 60);
    const seconds = progress.remainingTime % 60;
    document.getElementById("remaining").textContent = `${minutes
      .toString()
      .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  } else {
    document.getElementById("remaining").textContent = "--:--";
  }
}

function getStatusText(status) {
  const statusMap = {
    idle: "准备中",
    downloading: "下载中",
    completed: "已完成",
    error: "错误",
  };
  return statusMap[status] || status;
}

function resetProgress() {
  document.getElementById("progress-bar").style.width = "0%";
  document.getElementById("progress-text").textContent = "0%";
  document.getElementById("status").textContent = "准备中...";
  document.getElementById("status").className = "status";
  document.getElementById("speed").textContent = "0 MB/s";
  document.getElementById("downloaded").textContent = "0 MB";
  document.getElementById("total").textContent = "0 MB";
  document.getElementById("remaining").textContent = "--:--";
}
