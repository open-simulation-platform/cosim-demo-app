(ns cse-client.components
  (:require [cse-client.controller :as controller]
            [re-frame.core :as rf]
            [reagent.core :as r]))

(defn variable-override-editor [index {:keys [name type value-reference]} value]
  (let [editing? (r/atom false)
        edited? (r/atom false)
        internal-value (r/atom value)
        save (fn []
               (rf/dispatch [::controller/set-value (str index) type (str value-reference) @internal-value])
               (reset! editing? false)
               (reset! edited? true))
        save-if-changed (fn [value]
                          (if (not= value @internal-value)
                            (save)
                            (reset! editing? false)))]
    (fn [_ _ value]
      (if @editing?
        [:div.ui.action.input.fluid
         (if (= "Boolean" type)
           [:select.ui.dropdown
            {:value     @internal-value
             :on-change #(reset! internal-value (.. % -target -value))}
            [:option {:value "false"} "false"]
            [:option {:value "true"} "true"]]
           [:input {:type         :text
                    :auto-focus   true
                    :id           (str "input-" name)
                    :value        @internal-value
                    :on-change    #(reset! internal-value (.. % -target -value))
                    :on-key-press #(when (= (.-key %) "Enter") (save-if-changed value))
                    ;:on-blur      #(save-if-changed value)
                    }])
         [:button.ui.right.icon.button
          {:on-click save}
          [:i.check.link.icon]]
         [:button.ui.right.icon.button
          {:on-click #(reset! editing? false)}
          [:i.times.link.icon]]]
        [:div
         [:span.plotname-edit
          {:on-click     (fn []
                           (reset! edited? false)
                           (rf/dispatch [::controller/reset-value (str index) type (str value-reference)]))
           :data-tooltip "Remove override"}
          [:i.eraser.icon]]
         [:span.plotname-edit
          {:on-click     (fn []
                           (reset! editing? true)
                           (reset! internal-value value))
           :data-tooltip "Override value"}
          [:i.edit.link.icon]]
         (if (and @edited? (= "pause" @(rf/subscribe [:status])))
           @internal-value
           (str value))]))))

(defn text-editor [value event tooltip]
  (let [editing? (r/atom false)
        internal-value (r/atom value)
        save (fn []
               (rf/dispatch (conj event @internal-value))
               (reset! editing? false))
        save-if-changed (fn [value]
                          (if (not= value @internal-value)
                            (save)
                            (reset! editing? false)))]
    (fn [value]
      (if @editing?
        [:div.ui.action.input.fluid
         [:input {:type         :text
                  :auto-focus   true
                  :id           (str "input-" name)
                  :value        @internal-value
                  :on-change    #(reset! internal-value (.. % -target -value))
                  :on-key-press #(when (= (.-key %) "Enter") (save-if-changed value))
                  :on-blur      #(save-if-changed value)}]
         [:button.ui.right.icon.button
          {:on-click save}
          [:i.check.link.icon]]
         [:button.ui.right.icon.button
          {:on-click #(reset! editing? false)}
          [:i.times.link.icon]]]
        [:div
         [:span.plotname-edit
          {:on-click     (fn []
                           (reset! editing? true)
                           (reset! internal-value value))
           :data-tooltip tooltip}
          [:i.edit.link.icon]]
         value]))))
