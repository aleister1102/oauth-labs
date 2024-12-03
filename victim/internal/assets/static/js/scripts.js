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
            icon: 'small exclamation triangle icon',
            content: 'icon content',
        }
    });
}

function showMessage(message) {
    $.toast({
        message: message,
        position: 'top center',
        displayTime: 5000,
        closeIcon: false,
        class: 'info',
        preserveHTML: false,
        newestOnTop: true,
        showIcon: true,
        className: {
            toast: 'ui toast',
            icon: 'small exclamation circle icon',
            content: 'icon content',
        }
    });
}

$(document).ready(function() {
    $('form div.dropdown').dropdown();
});
