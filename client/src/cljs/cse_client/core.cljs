(ns cse-client.core
  (:require [cljsjs.bootstrap]
            [cse-client.routes :as routes]
            [re-frame.core :refer [dispatch subscribe dispatch-sync]]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [ajax.core :as ajax]))

(enable-console-print!)

(def default-db {})

(k/reg-controller :rest-demo-controller
                  {:params #(when (-> % :handler (= :rest-demo)) true)
                   :start  [:load-rest-demo]})

(k/reg-chain :load-rest-demo
             (fn [_ _]
               {:http-xhrio {:method          :get
                             :uri             "/rest-test"
                             :response-format (ajax/json-response-format)}})
             (fn [_ [_ rest-of-it]]
               (js/alert (str "GOT ME SOME REST from \"/rest-test\": " rest-of-it))))

(defn root-comp []
  (let [route (rf/subscribe [:kee-frame/route])]
    (fn []
      [:ul
       [:li [:a {:href (k/path-for [:index])} "Index"]]
       [:li [:a {:href (k/path-for [:sub1])} "sub1"]]
       [:li [:a {:href (k/path-for [:rest-demo])} "This one is real and will load the REST"]]])))

(k/start! {:routes         routes/routes
           :hash-routing?  true
           :debug?         true
           :root-component [root-comp]
           :initial-db     default-db})