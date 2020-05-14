(ns cse-client.localstorage
  (:require [oops.core :refer [ocall oget]]))

(def local-storage
  (oget js/window "localStorage"))

(defn set-item!
  [key val]
  (ocall local-storage "setItem" key val))

(defn get-item
  [key]
  (ocall local-storage "getItem" key))

(defn remove-item!
  [key]
  (ocall local-storage "removeItem" key))