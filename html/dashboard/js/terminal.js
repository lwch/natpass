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
        $.get('/api/rules', function(ret) {
            $('#terms').empty();
            $.each(ret, function(_, rule) {
                if (rule.type != 'shell' &&
                    rule.type != 'vnc' &&
                    rule.type != 'code-server') {
                    return;
                }
                $('#terms').append($(`<option value="${rule.port}">${rule.name}</option>`));
            });
            cb();
        });
    },
    connect: function() {
        $('#tabs>.nav-item>.active').removeClass('active');
        $('#tab-content>.active').removeClass('show').removeClass('active');
        var idx = page.idx;
        var str = `
        <li class="nav-item" id="tab-${idx}">
            <button class="nav-link active" type="button">
                shell - [${$('#terms option:selected').text()}]
            </button>
        </li>`;
        $('#tabs').append($(str));
        $("#tabs #tab-${idx}").click(function() {
            var $this = $(this);
            $('#tabs>.nav-item>.active').removeClass('active');
            $this.find('button').addClass('active');
            $('#tab-content>.active').removeClass('show').removeClass('active');
            $('#tab-content-'+idx).addClass('show').addClass('active');
        });
        var str = `
        <div class="tab-pane fade show active" id="tab-content-${idx}">
            <iframe src="http://${location.hostname}:${$('#terms').val()}"
                    style="width:100%;min-height:800px" allowfullscreen></iframe>
        </div>`;
        $('#tab-content').append($(str));
        page.idx++;
    },
    idx: 0
};
$(document).ready(page.init);