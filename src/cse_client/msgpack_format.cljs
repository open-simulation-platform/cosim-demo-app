(ns cse-client.msgpack-format
  (:require [chord.format :as f]
            [cse-client.msgpack :as msgpack]))

(defn unpack [s]
  (-> s
      msgpack/unpack
      clojure.walk/keywordize-keys))

(defn pack [obj]
  (-> obj
      clojure.walk/stringify-keys
      msgpack/pack))

(defmethod f/formatter* :msgpack [_]
  (reify f/ChordFormatter

    (freeze [_ obj]
      (pack obj))

    (thaw [_ s]
      (unpack s))))
