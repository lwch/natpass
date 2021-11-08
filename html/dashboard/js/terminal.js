var page = {
    init: function() {
        page.load(function() {
            var name = arg('name');
            if (name != 'null') {
                $('#terms').find(`option:contains(${name})`).prop('selected', true);
                page.connect();
            }
        });
        $('#connect').click(page.connect);
    },
    load: function(cb) {
        if (!cb) cb = function(){};
        $.get('/api/tunnels', function(ret) {
            $('#terms').empty();
            $.each(ret, function(_, tunnel) {
                if (tunnel.type != 'shell' &&
                    tunnel.type != 'vnc') {
                    return;
                }
                $('#terms').append($(`<option value="${tunnel.port}">${tunnel.name}</option>`));
            });
            cb();
        });
    },
    connect: function() {
        $('#tabs>.nav-item>.active').removeClass('active');
        $('#tab-content>.active').removeClass('show').removeClass('active');
        var idx = page.idx;
        var str = `
        <li class="nav-item">
            <button class="nav-link active" type="button">
                shell - [${$('#terms option:selected').text()}]
            </button>
        </li>`;
        var obj = $(str);
        obj.click(function() {
            var $this = $(this);
            $('#tabs>.nav-item>.active').removeClass('active');
            $this.find('button').addClass('active');
            $('#tab-content>.active').removeClass('show').removeClass('active');
            $('#tab-'+idx).addClass('show').addClass('active');
        });
        $('#tabs').append(obj);
        var str = `
        <div class="tab-pane fade show active" id="tab-${idx}">
            <iframe src="http://${location.hostname}:${$('#terms').val()}"
                    style="width:100%;min-height:800px" allowfullscreen></iframe>
        </div>`;
        var obj = $(str);
        $('#tab-content').append(obj);
        page.idx++;
    },
    idx: 0
};
$(document).ready(page.init);