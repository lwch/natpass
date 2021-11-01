var page = {
    init: function() {
        page.canvas = document.getElementById('vnc');
        page.ctx = page.canvas.getContext('2d');
        page.connect();
    },
    connect: function() {
        $.get('/new?quality=50', function(ret) {
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
    canvas: undefined,
    ctx: undefined,
    id: '',
    ws: undefined
};
$(document).ready(page.init);