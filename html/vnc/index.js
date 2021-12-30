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
        $('#vnc').contextmenu(page.right_click);
        $('#vnc').keydown(page.keydown);
        $('#vnc').keyup(page.keyup);
        $('#vnc').mousewheel(page.wheel);
        $('#quality').change(page.ctrl);
        $('#show-cursor').change(page.ctrl);
        $('#cad').click(page.cad);
        $('#clipboard').click(clipboard_dialog.modal);
        $('#fullscreen').click(page.fullscreen);
        // connect
        page.connect();
        setInterval(page.update_info, page.secs*1000);
    },
    connect: function() {
        var quality = $('#quality').val();
        var show_cursor = $('#show-cursor').prop('checked');
        $.get('/new?quality='+quality+'&show_cursor='+show_cursor, function(ret) {
            page.id = ret;
            page.ws = new WebSocket('ws://'+location.host+'/ws/'+ret);
            page.ws.onmessage = page.render;
            page.ws.onclose = page.closed;
        });
    },
    render: function(e) {
        page.fps++;
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
            var size = dv.getUint32(24, false);
            page.bandwidth += size;
            var id = page.ctx.getImageData(dx, dy, dwidth, dheight);
            var data = new Uint8Array(this.result.slice(28));
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
        if (!page.ws) {
            return;
        }
        page.ws.send(JSON.stringify({action: 'cad'}));
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
    right_click: function(e) {
        if (!page.ws) {
            return;
        }
        var pointer = page.get_pointer(e);
        pointer.button = 'right';
        pointer.status = 'down';
        page.ws.send(JSON.stringify({
            action: 'mouse',
            payload: pointer
        }));
        pointer.status = 'up';
        page.ws.send(JSON.stringify({
            action: 'mouse',
            payload: pointer
        }));
        return false;
    },
    get_pointer: function(e) {
        var width = parseInt($(e.target).css('width'), 10);
        var height = parseInt($(e.target).css('height'), 10);
        var x = e.offsetX - ((width - page.canvas.width) >> 1);
        var y = e.offsetY - ((height - page.canvas.height) >> 1);
        // page.ctx.beginPath();
        // page.ctx.arc(x, y, 5, 0, 2*Math.PI);
        // page.ctx.stroke();
        return {
            x: x,
            y: y
        }
    },
    keydown: function(e) {
        page.keyboard(e, 'down');
        return false;
    },
    keyup: function(e) {
        page.keyboard(e, 'up');
        return false;
    },
    keyboard: function(e, status) {
        if (!page.ws) {
            return;
        }
        var key = '';
        if ((e.which >= 65 && e.which <= 90) || // a-z
            (e.which >= 48 && e.which <= 57)) { // 0-9
            key = String.fromCharCode(e.which).toLowerCase();
        } else if (e.which >= 112 && e.which <= 123) {
            key = 'f'+(e.which - 111);
        } else if (e.which == 8) {
            key = 'backspace';
        } else if (e.which == 46) {
            key = 'delete';
        } else if (e.which == 13) {
            key = 'enter';
        } else if (e.which == 9) {
            key = 'tab';
        } else if (e.which == 27) {
            key = 'esc';
        } else if (e.which == 38) {
            key = 'up';
        } else if (e.which == 40) {
            key = 'down';
        } else if (e.which == 39) {
            key = 'right';
        } else if (e.which == 37) {
            key = 'left';
        } else if (e.which == 36) {
            key = 'home';
        } else if (e.which == 35) {
            key = 'end';
        } else if (e.which == 33) {
            key = 'pageup';
        } else if (e.which == 34) {
            key = 'pagedown';
        } else if (e.which == 16) {
            key = 'shift';
        } else if (e.which == 17) {
            key = 'control';
        } else if (e.which == 18) {
            key = 'alt';
        } else if (e.which == 32) {
            key = 'space';
        } else if (e.which == 189) {
            key = '-';
        } else if (e.which == 187) {
            key = '=';
        } else if (e.which == 219) {
            key = '[';
        } else if (e.which == 221) {
            key = ']';
        } else if (e.which == 220) {
            key = '\\';
        } else if (e.which == 186) {
            key = ';';
        } else if (e.which == 222) {
            key = "'";
        } else if (e.which == 188) {
            key = ',';
        } else if (e.which == 190) {
            key = '.';
        } else if (e.which == 191) {
            key = '/';
        } else if (e.which == 91) {
            key = 'cmd';
        }
        page.ws.send(JSON.stringify({
            action: 'keyboard',
            payload: {
                key: key,
                status: status
            }
        }));
        console.log(e);
    },
    update_info: function() {
        var str = 'fps: '+parseInt(page.fps/page.secs)+
            ', bandwidth: '+humanize.bytes(page.bandwidth/page.secs)+'/s';
        page.fps = 0;
        page.bandwidth = 0;
        $('#info').text(str);
    },
    closed: function() {
        $('#closed').css('display', 'inline-block');
        $('#info').css('display', 'none');
    },
    fullscreen: function() {
        $('#vnc')[0].requestFullscreen();
    },
    wheel: function(e) {
        page.ws.send(JSON.stringify({
            action: 'scroll',
            payload: {
                x: parseInt(e.deltaX),
                y: parseInt(e.deltaY)
            }
        }));
        return false;
    },
    canvas: undefined,
    ctx: undefined,
    id: '',
    ws: undefined,
    secs: 2,
    fps: 0,
    bandwidth: 0
};
$(document).ready(page.init);