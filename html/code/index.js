var page = {
    init: function() {
        page.connect();
    },
    connect: function() {
        $.get('/new', function(ret) {
            page.id = ret.id;
            page.name = ret.name;
            $('#code').attr('src', `/forward/${page.name}/?id=${page.id}`);
        });
    },
    id: '',
    name: ''
};
$(document).ready(page.init);