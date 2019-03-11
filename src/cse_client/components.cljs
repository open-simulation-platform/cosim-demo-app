(ns cse-client.components
  (:require [cse-client.controller :as controller]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [reagent.core :as r]))

(defn variable-override-editor [module {:keys [name causality type]} value event]
  (let [editing? (r/atom false)
        internal-value (r/atom value)
        save (fn []
               (rf/dispatch (if event
                              (conj event @internal-value)
                              [::controller/set-value module name causality type @internal-value]))
               (reset! editing? false))
        save-if-changed (fn [value]
                          (if (not= value @internal-value)
                              (save)
                              (reset! editing? false)))]
    (fn [_ _ value]
      (if @editing?
        [:div.ui.action.input.fluid
         [:input {:type      :text
                  :auto-focus true
                  :id        (str "input-" name)
                  :value     (if @editing? @internal-value value)
                  :on-change #(reset! internal-value (.. % -target -value))
                  :on-key-press #(when (= (.-key %) "Enter") (save-if-changed value))
                  :on-blur #(save-if-changed value)}]
         [:button.ui.right.icon.button
          {:on-click save}
          [:i.check.link.icon]]
         [:button.ui.right.icon.button
          {:on-click #(reset! editing? false)}
          [:i.times.link.icon]]]

        [:div.plotname-edit
         {:on-click (fn []
                      (reset! editing? true)
                      (reset! internal-value value))}

         [:span [:i.edit.link.icon]] value]))))
