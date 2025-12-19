import { BarChart3, TrendingUp, Calendar, Target, Info, ExternalLink } from 'lucide-react';

function AboutPage() {
  return (
    <main className="mx-auto px-2 sm:px-4 py-8">
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Application Title and Description */}
        <section className="text-center">
          <h2 className="text-3xl font-bold text-gray-900 mb-4">About WaniKani Dashboard</h2>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            A comprehensive progress tracking dashboard for WaniKani learners, providing detailed insights 
            into your Japanese kanji and vocabulary learning journey.
          </p>
        </section>

        {/* What is WaniKani */}
        <section className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center gap-3 mb-4">
            <Info className="w-6 h-6 text-blue-600" />
            <h3 className="text-xl font-semibold text-gray-900">What is WaniKani?</h3>
          </div>
          <p className="text-gray-700 mb-4">
            WaniKani is a spaced repetition system (SRS) designed to help you learn Japanese kanji and vocabulary 
            efficiently. It uses a systematic approach to introduce new characters and reinforce your knowledge 
            through timed reviews.
          </p>
          <p className="text-gray-700 mb-4">
            This dashboard helps you track your progress by visualizing your learning data, showing you patterns 
            in your study habits, and helping you understand how well you're retaining the material you've learned.
          </p>
          <a 
            href="https://www.wanikani.com" 
            target="_blank" 
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 text-blue-600 hover:text-blue-800 font-medium"
          >
            Visit WaniKani <ExternalLink className="w-4 h-4" />
          </a>
        </section>

        {/* Dashboard Features */}
        <section className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center gap-3 mb-6">
            <BarChart3 className="w-6 h-6 text-blue-600" />
            <h3 className="text-xl font-semibold text-gray-900">Dashboard Features</h3>
          </div>
          
          <div className="grid md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <Target className="w-5 h-5 text-green-600 mt-1 flex-shrink-0" />
                <div>
                  <h4 className="font-medium text-gray-900 mb-1">Item Spread Overview</h4>
                  <p className="text-sm text-gray-600">
                    View the distribution of your learned items across different SRS stages, 
                    from Apprentice to Burned items.
                  </p>
                </div>
              </div>
              
              <div className="flex items-start gap-3">
                <Calendar className="w-5 h-5 text-purple-600 mt-1 flex-shrink-0" />
                <div>
                  <h4 className="font-medium text-gray-900 mb-1">Daily Totals Chart</h4>
                  <p className="text-sm text-gray-600">
                    Track your daily review activity and see how many items you've 
                    reviewed at each SRS level over time.
                  </p>
                </div>
              </div>
            </div>
            
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <TrendingUp className="w-5 h-5 text-orange-600 mt-1 flex-shrink-0" />
                <div>
                  <h4 className="font-medium text-gray-900 mb-1">Daily Proportions Chart</h4>
                  <p className="text-sm text-gray-600">
                    Analyze the percentage breakdown of your reviews by SRS stage, 
                    helping you understand your learning patterns.
                  </p>
                </div>
              </div>
              
              <div className="flex items-start gap-3">
                <BarChart3 className="w-5 h-5 text-blue-600 mt-1 flex-shrink-0" />
                <div>
                  <h4 className="font-medium text-gray-900 mb-1">Progress Visualization</h4>
                  <p className="text-sm text-gray-600">
                    Interactive charts that help you identify trends, spot areas 
                    for improvement, and celebrate your progress.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Chart Interpretation Guide */}
        <section className="bg-white rounded-lg shadow-md p-6">
          <h3 className="text-xl font-semibold text-gray-900 mb-6">Understanding Your Charts</h3>
          
          <div className="space-y-6">
            <div className="border-l-4 border-blue-500 pl-4">
              <h4 className="font-medium text-gray-900 mb-2">SRS Stages Explained</h4>
              <div className="grid sm:grid-cols-2 gap-3 text-sm">
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="font-medium text-pink-600">Apprentice:</span>
                    <span className="text-gray-600">Learning new items (4 hours - 1 week)</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="font-medium text-purple-600">Guru:</span>
                    <span className="text-gray-600">Getting familiar (2 weeks - 1 month)</span>
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="font-medium text-blue-600">Master:</span>
                    <span className="text-gray-600">Well-known items (4 months)</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="font-medium text-yellow-600">Enlightened:</span>
                    <span className="text-gray-600">Nearly mastered (4 months)</span>
                  </div>
                </div>
              </div>
              <div className="mt-2">
                <div className="flex justify-between text-sm">
                  <span className="font-medium text-gray-800">Burned:</span>
                  <span className="text-gray-600">Permanently learned - no more reviews needed!</span>
                </div>
              </div>
            </div>

            <div className="border-l-4 border-green-500 pl-4">
              <h4 className="font-medium text-gray-900 mb-2">Reading Your Progress</h4>
              <ul className="text-sm text-gray-700 space-y-1">
                <li>• <strong>High Apprentice numbers:</strong> You're learning lots of new material</li>
                <li>• <strong>Growing Guru items:</strong> Your knowledge is solidifying</li>
                <li>• <strong>Increasing Burned items:</strong> You're making permanent progress</li>
                <li>• <strong>Daily consistency:</strong> Regular study habits lead to better retention</li>
              </ul>
            </div>

            <div className="border-l-4 border-orange-500 pl-4">
              <h4 className="font-medium text-gray-900 mb-2">Tips for Success</h4>
              <ul className="text-sm text-gray-700 space-y-1">
                <li>• Aim for consistent daily reviews rather than cramming</li>
                <li>• Don't worry if Apprentice items pile up - it's normal when learning actively</li>
                <li>• Focus on accuracy over speed to build strong foundations</li>
                <li>• Use the charts to identify your most productive study times</li>
              </ul>
            </div>
          </div>
        </section>

        {/* Footer */}
        <section className="text-center text-gray-500 text-sm border-t pt-6">
          <p>
            This dashboard is designed to complement your WaniKani learning experience. 
            Keep up the great work on your Japanese learning journey!
          </p>
        </section>
      </div>
    </main>
  );
}

export default AboutPage;