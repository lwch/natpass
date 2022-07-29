function arg(key) {
    var url = new URL(location.href);
    return decodeURIComponent(url.searchParams.get(key));
}
function escape(htmlStr) {
    return htmlStr.replace(/&/g, "&amp;")
          .replace(/</g, "&lt;")
          .replace(/>/g, "&gt;")
          .replace(/"/g, "&quot;")
          .replace(/'/g, "&#39;");
}
var humanize = {
    bytes: function(n) {
        return humanize.humanate_bytes(n, 1024, ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'])
    },
    humanate_bytes: function(n, base, sizes) {
        if (n < 10) {
            return n+'B';
        }
        var e = Math.floor(humanize.logn(n, base));
        var suffix = sizes[e];
        var val = Math.floor(n/Math.pow(base, e)*10+0.5) / 10
        return val.toFixed(1)+suffix;
    },
    logn: function(n, base) {
        return Math.log(n) / Math.log(base)
    }
};