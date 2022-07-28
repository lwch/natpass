var page = {
    init: function() {
        $('#fullscreen').click(page.fullscreen);
        $('#code').on('load', function() {
            var qry = $('#code')[0].contentWindow.location.search;
            var params = new URLSearchParams(qry);
            page.id = params.get('natpass_connection_id');
        });
        page.connect();
        setInterval(page.update_info, page.secs*1000);
    },
    connect: function() {
        $.get('/new', function(ret) {
            page.name = ret;
            $('#code').attr('src', `/forward/${page.name}/`);
        });
    },
    update_info: function() {
        if (!page.id) {
            return;
        }
        $.get('/info?id='+page.id, function(ret) {
            var send_bytes = ret.send_bytes - page.send;
            var recv_bytes = ret.recv_bytes - page.recv;
            if (send_bytes < 0) {
                send_bytes = 0;
            }
            if (recv_bytes < 0) {
                recv_bytes = 0;
            }
            page.send = ret.send_bytes;
            page.recv = ret.recv_bytes;
            var str = 'send: '+humanize.bytes(send_bytes/page.secs)+'/s, '+
                      'recv: '+humanize.bytes(recv_bytes/page.secs)+'/s';
            $('#info').text(str);
        });
    },
    fullscreen: function() {
        $('#code')[0].requestFullscreen();
    },
    id: '',
    name: '',
    secs: 2,
    send: 0,
    recv: 0
};
$(document).ready(page.init);