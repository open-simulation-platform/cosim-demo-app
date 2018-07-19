(ns cse-client.core
  (:require [cljsjs.bootstrap]
            [cse-client.routes :as routes]
            [re-frame.core :refer [dispatch subscribe dispatch-sync]]
            [kee-frame.core :as k]
            [day8.re-frame.http-fx]
            [bidi.bidi :as bidi]
            [kee-frame.api :as api]
            [clojure.string :as string]
            [re-frame.core :as rf]))

(enable-console-print!)

(def default-db {})

(defn root-comp []
  (let [route (rf/subscribe [:kee-frame/route])]
    (fn []
      [:ul
       [:li [:a {:href (k/path-for [:index])} "Index"]]
       [:li [:a {:href (k/path-for [:article])} "Article"]]])))

(k/start! {:routes         routes/routes
           :hash-routing?  true
           :debug?         true
           :root-component [root-comp]
           :initial-db     default-db})