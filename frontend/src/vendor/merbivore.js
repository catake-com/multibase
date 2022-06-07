ace.define("ace/theme/merbivore_custom", ["require", "exports", "module", "ace/lib/dom"], function(require, exports, module) {

  exports.isDark = true;
  exports.cssClass = "ace-merbivore-custom";
  exports.cssText = ".ace-merbivore-custom .ace_gutter {\
background: #262424;\
color: #E6E1DC\
}\
.ace-merbivore-custom .ace_print-margin {\
width: 1px;\
background: #262424\
}\
.ace-merbivore-custom {\
background-color: #1C1C1C;\
color: #E6E1DC\
}\
.ace-merbivore-custom .ace_cursor {\
color: #FFFFFF\
}\
.ace-merbivore-custom .ace_marker-layer .ace_selection {\
background: #494949\
}\
.ace-merbivore-custom.ace_multiselect .ace_selection.ace_start {\
box-shadow: 0 0 3px 0px #1C1C1C;\
}\
.ace-merbivore-custom .ace_marker-layer .ace_step {\
background: rgb(102, 82, 0)\
}\
.ace-merbivore-custom .ace_marker-layer .ace_bracket {\
margin: -1px 0 0 -1px;\
border: 1px solid #404040\
}\
.ace-merbivore-custom .ace_marker-layer .ace_active-line {\
background: #333435\
}\
.ace-merbivore-custom .ace_gutter-active-line {\
background-color: #333435\
}\
.ace-merbivore-custom .ace_marker-layer .ace_selected-word {\
border: 1px solid #494949\
}\
.ace-merbivore-custom .ace_invisible {\
color: #404040\
}\
.ace-merbivore-custom .ace_entity.ace_name.ace_tag,\
.ace-merbivore-custom .ace_keyword,\
.ace-merbivore-custom .ace_meta,\
.ace-merbivore-custom .ace_meta.ace_tag,\
.ace-merbivore-custom .ace_storage {\
color: #FC803A\
}\
.ace-merbivore-custom .ace_constant,\
.ace-merbivore-custom .ace_constant.ace_character,\
.ace-merbivore-custom .ace_constant.ace_character.ace_escape,\
.ace-merbivore-custom .ace_constant.ace_other,\
.ace-merbivore-custom .ace_support.ace_type {\
color: #68C1D8\
}\
.ace-merbivore-custom .ace_constant.ace_character.ace_escape {\
color: #B3E5B4\
}\
.ace-merbivore-custom .ace_constant.ace_language {\
color: #E1C582\
}\
.ace-merbivore-custom .ace_constant.ace_library,\
.ace-merbivore-custom .ace_string,\
.ace-merbivore-custom .ace_support.ace_constant {\
color: #8EC65F\
}\
.ace-merbivore-custom .ace_constant.ace_numeric {\
color: #7FC578\
}\
.ace-merbivore-custom .ace_invalid,\
.ace-merbivore-custom .ace_invalid.ace_deprecated {\
color: #FFFFFF;\
background-color: #FE3838\
}\
.ace-merbivore-custom .ace_fold {\
background-color: #FC803A;\
border-color: #E6E1DC\
}\
.ace-merbivore-custom .ace_comment,\
.ace-merbivore-custom .ace_meta {\
font-style: italic;\
color: #AC4BB8\
}\
.ace-merbivore-custom .ace_variable,\
.ace-merbivore-custom .ace_variable.ace_language {\
color: #8fbcbb;\
}\
.ace-merbivore-custom .ace_entity.ace_other.ace_attribute-name {\
color: #EAF1A3\
}\
.ace-merbivore-custom .ace_indent-guide {\
background: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAACCAYAAACZgbYnAAAAEklEQVQImWOQkpLyZfD09PwPAAfYAnaStpHRAAAAAElFTkSuQmCC) right repeat-y\
}";

  var dom = require("../lib/dom");
  dom.importCssString(exports.cssText, exports.cssClass, false);
});
(function() {
  ace.require(["ace/theme/merbivore_custom"], function(m) {
    if (typeof module == "object" && typeof exports == "object" && module) {
      module.exports = m;
    }
  });
})();
