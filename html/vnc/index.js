var page = {
    init: function() {
        page.canvas = document.getElementById('vnc');
        page.ctx = page.canvas.getContext('2d');
        // bind events
        $('#vnc').bind('dragstart', function() {
            return false;
        });
        $('#vnc').mousemove(page.mousemove);
        $('#vnc').mousedown(page.mousedown);
        $('#vnc').mouseup(page.mouseup);
        $('#quality').change(page.ctrl);
        $('#show-cursor').change(page.ctrl);
        $('#cad').click(page.cad);
        // connect
        page.connect();
    },
    connect: function() {
        var quality = $('#quality').val();
        var show_cursor = $('#show-cursor').prop('checked');
        $.get('/new?quality='+quality+'&show_cursor='+show_cursor, function(ret) {
            page.id = ret;
            page.ws = new WebSocket('ws://'+location.host+'/ws/'+ret);
            page.ws.onmessage = page.render;
        });
    },
    render: function(e) {
        var reader = new FileReader;
        reader.onload = function() {
            var dv = new DataView(this.result);
            var screen_width = dv.getUint32(0, false);
            var screen_height = dv.getUint32(4, false);
            if (screen_width != page.canvas.width ||
                screen_height != page.canvas.height) {
                page.canvas.width = screen_width;
                page.canvas.height = screen_height;
            }
            var dx = dv.getUint32(8, false);
            var dy = dv.getUint32(12, false);
            var dwidth = dv.getUint32(16, false);
            var dheight = dv.getUint32(20, false);
            var id = page.ctx.getImageData(dx, dy, dwidth, dheight);
            var data = new Uint8Array(this.result.slice(24));
            var buf = id.data;
            for (var i = 0; i < buf.length; i++) {
                buf[i] = data[i];
                if (i % 4 == 3) {
                    buf[i] = 255; // alpha
                }
            }
            page.ctx.putImageData(id, dx, dy);
        };
        reader.readAsArrayBuffer(e.data);
    },
    ctrl: function() {
        var quality = $('#quality').val();
        var show_cursor = $('#show-cursor').prop('checked');
        $.post('/ctrl', {
            quality: quality,
            show_cursor: show_cursor
        });
    },
    cad: function() {
        // TODO
    },
    mousemove: function(e) {
        if (!page.ws) {
            return;
        }
        page.ws.send(JSON.stringify({
            action: 'mouse',
            payload: page.get_pointer(e)
        }));
    },
    mousedown: function(e) {
        if (!page.ws) {
            return;
        }
        var pointer = page.get_pointer(e);
        pointer.button = 'left';
        pointer.status = 'down';
        page.ws.send(JSON.stringify({
            action: 'mouse',
            payload: pointer
        }));
    },
    mouseup: function(e) {
        if (!page.ws) {
            return;
        }
        var pointer = page.get_pointer(e);
        pointer.button = 'left';
        pointer.status = 'up';
        page.ws.send(JSON.stringify({
            action: 'mouse',
            payload: pointer
        }));
    },
    get_pointer: function(e) {
        return {
            x: e.offsetX,
            y: e.offsetY
        }
    },
    canvas: undefined,
    ctx: undefined,
    id: '',
    ws: undefined
};
$(document).ready(page.init);