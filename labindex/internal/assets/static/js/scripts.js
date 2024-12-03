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
