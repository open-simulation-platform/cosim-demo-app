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

(defrecord BidiRouter [routes]
  api/Router
  (data->url [_ data]
    (str "/#" (apply bidi/path-for routes data)))
  (url->data [_ url]
    (let [[path+query fragment] (-> url (string/replace #"^/#" "") (string/split #"#" 2))
          [path query] (string/split path+query #"\?" 2)]
      (some-> (bidi/match-route routes path)
              (assoc :query-string query :hash fragment)))))

(defn root-comp []
  (let [route (rf/subscribe [:kee-frame/route])]
    (fn []
      [:ul
       [:li [:a {:href (k/path-for [:index])} "Index"]]
       [:li [:a {:href (k/path-for [:article])} "Article"]]])))

(k/start! {:router         (->BidiRouter routes/routes)
           :debug?         true
           :root-component [root-comp]
           :initial-db     default-db})