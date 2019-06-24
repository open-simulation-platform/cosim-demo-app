(ns cse-client.guide
  (:require [re-frame.core :as rf]
            [cse-client.controller :as controller]))

(def about-content [:div
                    [:p "This application demonstrates the features of the Core Simulation Environment, with basic
                    co-simulation functionality exposed through a simple graphical user interface. From this application
                    you will be able to:"]
                    [:ul
                     [:li "Load a configuration for co-simulation"]
                     [:li "Apply basic simulation control"]
                     [:li "Observe and trend any simulation variables"]
                     [:li "Log simulation variables to file"]]
                    [:p "The application is intended to be a generic tool for running simple co-simulations with the
                    cse-core simulation engine."]])

(def config-content [:div
                     [:h4 "Creating a co-simulation"]
                     [:p "The co-simulation should consist of one or more FMUs placed together in a folder. The path to this
                      folder needs to be provided in the \"FMUs\" input field on the simulation setup page. Optionally specify a path where
                      log files should be placed in the \"logs\" input field."]
                     [:h4 "Creating a SSP configuration file"]
                     [:p "For specifying connections between FMUs in a co-simulation, cse-core currently supports the SSP standard for FMUs.
                     An example SSP setup with two FMUs can be downloaded below."]
                     [:a {:target "TOP" :href "static/xml/ExampleSystemStructure.xml" :download ""} "Example SSP file"]
                     [:h4 "Creating a log configuration file"]
                     [:p "For configuring what signals to log in each FMU, cse-core optionally supports the inclusion of a log config XML file, to be put
                     in the same directory as the FMUs."]
                     [:i "The file must be named LogConfig.xml (including case)."]
                     [:p ""]
                     [:p "An example file is provided below."]
                     [:p "Note that while a config can be provided in one file for all the simulators in the simulation, output will still be to one file
                     pr. simulator / FMU. An optional attribute" [:i " decimationFactor"] " can be provided on the simulator level, which will specify how often
                     simulator values are to be logged (logging only every decimationFactor sample)."]
                     [:p "Finally, to log all signals for a simulator, simply leave the variable list empty as shown in the example file."]
                     [:a {:target "TOP" :href "static/xml/LogConfig.xml" :download ""} "Example logger configuration file"]])

(def simulation-content [:div
                         [:h4 "Running a simulation"]
                         [:p "Once the simulation has been loaded successfully, the FMUs will be listed in the left sidebar.
                         The individual FMUs can be browsed by clicking the names. To start the simulation, click the
                         \"Play\" button."]
                         [:h4 "Modifying and trending variables"]
                         [:p "To modify a variable value, click the \"value\" column in the variable table for that FMU."]
                         [:p "To trend a variable, click \"Add to trend\". To view the trend, click the \"Trend\" option in the
                         left sidebar. Currently all variables will be added to the same trend, but this will be improved upon in
                         future releases."]])

(def guide-data (array-map "About" about-content
                           "Configuration" config-content
                           "Simulation" simulation-content))

(defn form []
  (let [active @(rf/subscribe [:active-guide-tab])]
    [:div.ui.one.column.grid
     [:div.one.column.row
      [:div.column
       [:div.ui.top.attached.tabular.menu
        (map (fn [[header _]]
               ^{:key (str "guide-tab-" header)}
               [:a.item {:data-tab header
                         :class    (when (= header active) "active")
                         :on-click #(rf/dispatch [::controller/guide-navigate header])}
                header]) guide-data)]
       (map (fn [[header content]]
              ^{:key (str "guide-content-" header)}
              [:div.ui.bottom.attached.tab.segment {:data-tab header
                                                    :class    (when (= header active) "active")}
               content]) guide-data)]]]))