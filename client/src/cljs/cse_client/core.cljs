(ns cse-client.core
  (:require [cljsjs.bootstrap]
            [cse-client.routes :as routes]
            [re-frame.core :refer [dispatch subscribe dispatch-sync]]
            [kee-frame.core :as k]
            [re-frame.core :as rf]))

(enable-console-print!)

(def default-db {})

(defn root-comp []
  (let [route (rf/subscribe [:kee-frame/route])]
    (fn []
      [:ul
       [:li [:a {:href (k/path-for [:index])} "Index"]]
       [:li [:a {:href (k/path-for [:sub1])} "sub1"]]
       [:li [:a {:href (k/path-for [:sub2])} "sub2"]]])))

(k/start! {:routes         routes/routes
           :hash-routing?  true
           :debug?         true
           :root-component [root-comp]
           :initial-db     default-db})