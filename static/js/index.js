document.getElementById('searchBtn').addEventListener('click', function() {
    const username = document.getElementById('userInput').value.trim();
    if (!username) {
        alert('请输入用户名');
        return;
    }

    fetch(`/api/get_user/${encodeURIComponent(username)}`)
        .then(res => res.json())
        .then(res => {
            if (res.data) {
                const dataStr = encodeURIComponent(JSON.stringify(res.data));
                window.location.href = `/user_detail.html?data=${dataStr}`;
            } else {
                alert('未找到用户');
            }
        })
        .catch(err => {
            console.error(err);
            alert('请求失败');
        });
});
