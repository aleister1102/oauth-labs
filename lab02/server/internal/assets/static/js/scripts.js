function logout() {
    fetch('/logout', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        }
    }).then(() => {
        window.location.replace('/');
    })
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

$(document).ready(function() {
    $('#login-form').form({
        fields: {
            username: ['minLength[3]', 'maxLength[60]', 'empty'],
            password: ['maxLength[70]', 'empty'],
        }
    });
    $('#register-form').form({
        fields: {
            username: ['minLength[3]', 'maxLength[60]', 'empty'],
            password: ['maxLength[70]', 'empty'],
        }
    });
});
