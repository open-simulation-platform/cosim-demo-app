(ns cse-client.dp
  (:require [re-frame.core :as rf]))

(rf/reg-sub :boat (fn [db] {:x 50 :y 100 :rotation (-> db :state :module :signals first :value (or 0))}))

(defn boat []
  (let [{:keys [x y rotation]} @(rf/subscribe [:boat])]
    [:polygon {:transform (str "rotate(" rotation " 300 200)")
               :points    "100,300 100,50 300,50 300,300 200,400"
               :style     {:stroke       "black"
                           :fill         "white"
                           :stroke-width "3"}}]))

(defn svg-component []
  [:svg {:width  "720"
         :height "720"
         :id     "canvas"
         :style  {:outline          "1px solid black"
                  :background-color "#fee"}}
   [boat]])
