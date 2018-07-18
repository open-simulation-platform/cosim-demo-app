(ns cse-client.core
  (:require [cljsjs.bootstrap]
            [cse-client.routes :as routes]
            [re-frame.core :refer [dispatch subscribe dispatch-sync]]
            [kee-frame.core :as k]
            [day8.re-frame.http-fx]
            [bidi.bidi :as bidi]
            [kee-frame.api :as api]))

(enable-console-print!)

(def default-db {})

(defrecord BidiRouter [routes]
  api/Router
  (data->url [_ data]
    (apply bidi/path-for routes data))
  (url->data [_ url]
    (bidi/match-route routes url)))

(defn root-comp []
  [:div "MORDI IS HOME"])

(k/start! {:router         (->BidiRouter routes/routes)
           :debug?         true
           :root-component [root-comp]
           :initial-db     default-db})