var page = {
    init: function() {
        page.render();
        setInterval(page.render, 5000);
    },
    render: function() {
        page.render_cards();
        page.render_tunnels();
    },
    render_cards: function() {
        $.get('/api/info', function(ret) {
            $('#cards').empty();
            page.add_card('隧道总数', ret.tunnels);
            page.add_card('物理连接数', ret.physical_links);
            page.add_card('虚拟连接数', ret.virtual_links);
            page.add_card('终端会话', ret.sessions);
        });
    },
    render_tunnels: function() {
        $.get('/api/tunnels', function(ret) {
            $('#tunnels tbody').empty();
            $.each(ret, function(_, tunnel) {
                var send_bytes = 0;
                var send_packet = 0;
                var recv_bytes = 0;
                var recv_packet = 0;
                $.each(tunnel.links, function(_, link) {
                    send_bytes += link.send_bytes;
                    send_packet += link.send_packet;
                    recv_bytes += link.recv_bytes;
                    recv_packet += link.recv_packet;
                });
                var op = '';
                switch (tunnel.type) {
                case 'reverse':
                    op = `
                    <a href="http://${location.hostname}:${tunnel.port}" target="_blank">http</a>
                    <a href="https://${location.hostname}:${tunnel.port}" target="_blank">https</a>`;
                    break;
                case 'shell':
                    op = `<a href="http://${location.host}/terminal.html?name=${tunnel.name}" target="_blank">连接</a>`;
                    break;
                }
                var str = `
                <tr>
                    <td>${tunnel.name}</td>
                    <td>${tunnel.remote}</td>
                    <td>${tunnel.type}</td>
                    <td>${tunnel.links?tunnel.links.length:0}</td>
                    <td>${humanize.bytes(recv_bytes)}/${humanize.bytes(send_bytes)}</td>
                    <td>${recv_packet}/${send_packet}</td>
                    <td>${op}</td>
                </tr>`;
                $('#tunnels tbody').append($(str));
            });
        });
    },
    add_card: function(title, count) {
        var str = `
        <div class="col-lg-3">
            <div class="card">
                <div class="card-header border-0">
                    <div class="d-flex justify-content-between">
                        <h3 class="card-title">${title}</h3>
                    </div>
                </div>
                <div class="card-body">
                    <div class="d-flex">
                        <p class="d-flex flex-column">
                            <span class="text-bold text-lg">${count}</span>
                        </p>
                    </div>
                </div>
            </div>
        </div>`;
        $('#cards').append($(str));
    }
};
$(document).ready(page.init);