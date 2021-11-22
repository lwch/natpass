var clipboard_dialog = {
    init: function() {
        $('#dlg-clipboard #set-clipboard').click(clipboard_dialog.set);
        $('#dlg-clipboard #get-clipboard').click(clipboard_dialog.get);
    },
    modal: function() {
        $('#dlg-clipboard textarea').val('');
        $('#dlg-clipboard').modal();
    },
    set: function() {
        $.post('/clipboard', {data: $('#dlg-clipboard textarea').val()});
    },
    get: function() {
        $.get('/clipboard', function(ret) {
            $('#dlg-clipboard textarea').val(ret);
        });
    }
};
$(document).ready(clipboard_dialog.init);