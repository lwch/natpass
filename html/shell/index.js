var page = {
    init: function() {
        page.terminal = new Terminal({
            renderType: 'canvas'
        });
        page.terminal.open(document.getElementById('terminal'));
        page.terminal.writeln('正在连接...');
        $.get('/new', function(ret) {
            page.id = ret;
            page.websocket = new WebSocket('ws://'+location.host+'/ws/'+ret);
            page.websocket.onclose = page.onclose;
            page.terminal.reset();
            page.terminal.loadAddon(new AttachAddon.AttachAddon(page.websocket));
            document.getElementById('terminal').style.height = (window.innerHeight-1) + 'px';
            var fit = new FitAddon.FitAddon();
            page.terminal.loadAddon(fit);
            fit.fit();
            page.resize();
        });
    },
    resize: function() {
        $.post('/resize', {
            id: page.id,
            rows: page.terminal.rows,
            cols: page.terminal.cols
        });
    },
    onclose: function() {
        page.terminal.writeln('');
        page.terminal.writeln("\033[0;31m连接已断开！");
    },
    id: undefined,
    terminal: undefined,
    websocket: undefined
};
$(document).ready(page.init);