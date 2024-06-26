var DocumentRow = Backbone.View.extend({

    tagName: "div",

    className: "authfoo",

    events: {
        "click .icon":          "open",
        "click .button.edit":   "openEditDialog",
        "click .button.delete": "destroy"
    },

    initialize: function() {
        this.listenTo(this.model, "change", this.render);
    },

    render: function() {
    ...
    }

});