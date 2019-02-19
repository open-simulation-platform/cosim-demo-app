(ns cse-client.components
  (:require [cse-client.controller :as controller]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [reagent.core :as r]))

(defn variable-override-editor [module {:keys [name causality type]} value event]
  (let [editing? (r/atom false)
        internal-value (r/atom value)]
    (fn [_ _ value]
      (if @editing?
        [:div.ui.action.input.fluid
         [:input {:type      :text
                  :autoFocus true
                  :id        (str "input-" name)
                  :value     (if @editing? @internal-value value)
                  :on-change #(reset! internal-value (.. % -target -value))}]
         [:button.ui.right.icon.button
          {:on-click (fn [_]
                       (rf/dispatch (if event
                                      (conj event @internal-value)
                                      [::controller/set-value module name causality type @internal-value]))
                       (reset! editing? false))}
          [:i.check.link.icon]]
         [:button.ui.right.icon.button
          {:on-click #(reset! editing? false)}
          [:i.times.link.icon]]]
        [:div {:style    {:cursor :pointer}
               :on-click (fn [_]
                           (reset! editing? true)
                           (reset! internal-value value))}
         value]))))
