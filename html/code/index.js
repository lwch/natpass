var page = {
    init: function() {
        page.connect();
    },
    connect: function() {
        $.get('/new', function(ret) {
            page.name = ret;
            $('#code').attr('src', `/forward/${page.name}/`);
        });
    },
    name: ''
};
$(document).ready(page.init);