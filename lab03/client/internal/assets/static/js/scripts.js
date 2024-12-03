function logout() {
    fetch('/logout', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        }
    }).then(() => {
        window.location.replace('/');
    });
}

function showError(message) {
    $.toast({
        message: message,
        position: 'top center',
        displayTime: 5000,
        closeIcon: false,
        class: 'error',
        preserveHTML: false,
        newestOnTop: true,
        showIcon: true,
        className: {
            toast: 'ui toast',
            icon: 'ui small exclamation triangle icon',
            content: 'icon content',
        }
    });
}
