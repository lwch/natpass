var page = {
    init: function() {
        page.connect();
    },
    connect: function() {
        $.get('/new', function(ret) {
            page.id = ret;
            $('#code').attr('src', `/forward/${page.id}/`);
        });
    },
    id: ''
};
$(document).ready(page.init);